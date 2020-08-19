package app

import (
	"log"
	"sync"
)

type IceetimeService struct {
	TorrentDAO          TorrentDAO
	TorrentClient       TorrentClient
	IndexerQueryHandler IndexerQueryHandler
	Qualities           []Quality
	MinSeeders          int
}

func (s *IceetimeService) FindLocallyOrFetchMovie(imdbID string, title string, year string, minQualityIndex int) (*Torrent, *Error) {
	torrentDAO := s.TorrentDAO
	torrentClient := s.TorrentClient
	iqh := s.IndexerQueryHandler

	// Attempt to get a matching torrent on disk
	torrentOnDisk, err := torrentDAO.GetByImdbIDAndMinQuality(imdbID, 0)
	if err != nil {
		return nil, NewError(err, 500, LocalDBQueryErr)

	}
	if torrentOnDisk != nil {
		log.Println("INFO: Torrent search cache hit")
		return torrentOnDisk, nil
	}

	// Fetch torrent online
	torrents, terr := iqh.QueryMovie(imdbID, title, year, 1)
	if terr != nil {
		return nil, terr
	}
	if len(torrents) == 0 {
		return nil, NewError(nil, 404, IndexerQueryNoResultsErr)
	}

	torrent, terr := s.getBestTorrentFromIndexerQuery(torrents, s.Qualities[minQualityIndex])
	if torrent == nil {
		return nil, NewError(nil, 404, IndexerQueryNoResultsErr)
	}

	err = torrentClient.AddFromInfoHash(torrent.InfoHash)
	if terr != nil {
		return nil, NewError(err, 500, "could not add torrent from hash")
	}

	// Save torrent to disk/cache
	if err := torrentDAO.Save(torrent); err != nil {
		err := torrentClient.RemoveByHash(torrent.InfoHash)
		if err != nil {
			log.Println("BRUTAL: Could not remove torrent after attempting an add. This is super bad!")
		}
		return nil, NewError(err, 500, LocalDBSaveErr)
	}

	return torrent, nil
}

// getBestTorrentFromIndexerQuery goes through each torrent, adds it to the client, and get's metadata to make an educated
// decision on each torrent. It always removes the torrents after it is done so they need to be added afterwards.
// TODO: Refactor this!
func (s *IceetimeService) getBestTorrentFromIndexerQuery(torrents []Torrent, q Quality) (*Torrent, *Error) {
	bestScore := 0.0
	var bestTorrent *Torrent = nil
	mux := &sync.Mutex{}
	var wg sync.WaitGroup

	for _, t := range torrents {
		wg.Add(1)
		go func(t Torrent) {
			defer wg.Done()
			score := 0.0
			if t.Size < q.MinSize || t.Size > q.MaxSize {
				log.Printf("INFO: Passing on release %s because size %v is not correct.", t.Title, t.Size)
				return
			}
			if t.Seeders < s.MinSeeders {
				log.Printf("INFO: Passing on release %s because seeders: %v is less than minimum: %v", t.Title, t.Seeders, s.MinSeeders)
				return
			}

			// Add to client to get hash
			hash, err := s.TorrentClient.AddFromURLUknownScheme(t.Link, t.LinkAuth)
			defer s.TorrentClient.RemoveByHash(hash)
			if err != nil {
				log.Printf("WARNING: could not add torrent magnet for %s\n Err: %s", t.Title, err)
				return
			}
			t.InfoHash = hash

			// Attempt to find a valid file
			index, terr := s.getValidFileInTorrent(t)
			if terr != nil {
				log.Printf("INFO: Passing on release %s because there was no valid file.", t.Title)
				return
			}

			// TODO: Do better logic here to find "main" file
			t.MainFileIndex = index

			score += float64(t.Seeders) / 10

			if score > bestScore {
				copyT := t // Why does this object need to be copied?? So weird..
				mux.Lock()
				bestScore = score
				bestTorrent = &copyT
				mux.Unlock()
			}

		}(t)
	}

	wg.Wait()

	return bestTorrent, nil
}

func (s *IceetimeService) getValidFileInTorrent(t Torrent) (int, error) {
	// Get correct file
	files, err := s.TorrentClient.GetFiles(t.InfoHash)
	if err != nil {
		return 0, NewError(err, 500, "unable to get files for torrent")
	}

	for i, file := range files {
		if StringEndsInAny(file, GetSupportedVideoFileFormats()) && !StringContainsAnyOf(file, GetBlacklistedFileNameContents()) {
			return i, nil
		}
	}
	return 0, NewError(nil, 400, InvalidTorrentErr)
}
