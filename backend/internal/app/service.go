package app

import (
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/anacrolix/torrent"
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

	torrent, terr := s.getBestTorrentFromIndexerQueryAndAddToClient(torrents, s.Qualities[minQualityIndex])
	if torrent == nil {
		return nil, NewError(nil, terr.Code, IndexerQueryNoResultsErr)
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

func (s *IceetimeService) GetTorrentByID(id int) (*Torrent, *Error) {
	torrent, err := s.TorrentDAO.GetByID(int(id))
	if err != nil {
		return nil, NewError(err, http.StatusInternalServerError, LocalDBQueryErr)
	}
	if torrent == nil {
		return nil, NewError(err, http.StatusNotFound, TorrentByIDNotFoundErr)
	}

	return torrent, nil
}

func (s *IceetimeService) GetFileReaderForFileInTorrent(t *Torrent, fileIndex int) (torrent.Reader, *Error) {
	reader, err := s.TorrentClient.GetReaderForFileInTorrent(t.InfoHash, fileIndex)
	if err != nil {
		return nil, NewError(err, http.StatusInternalServerError, TorrentFileReaderErr)
	}

	return reader, nil
}

// getBestTorrentFromIndexerQueryAndAddToClient goes through each torrent, adds it to the client, and get's metadata to make an educated
// decision on each torrent. It always removes the torrents after it is done so they need to be added afterwards.
// TODO: Refactor this!
func (s *IceetimeService) getBestTorrentFromIndexerQueryAndAddToClient(torrents []Torrent, q Quality) (*Torrent, *Error) {
	// sort torrents by seeders (to get most available torrents first)
	sort.Slice(torrents, func(i, j int) bool {
		return torrents[i].Seeders > torrents[j].Seeders
	})

	for _, t := range torrents {
		if t.Size < q.MinSize || t.Size > q.MaxSize {
			log.Printf("INFO: Passing on release %s because size %v is not correct.", t.Title, t.Size)
			continue
		}
		if t.Seeders < s.MinSeeders {
			log.Printf("INFO: Passing on release %s because seeders: %v is less than minimum: %v", t.Title, t.Seeders, s.MinSeeders)
			continue
		}
		if StringContainsAnyOf(strings.ToLower(t.Title), GetBlacklistedTorrentNameContents()) {
			log.Printf("INFO: Passing on release %s because title contains one of these blacklisted words: %+v", t.Title, GetBlacklistedTorrentNameContents())
			continue
		}

		// Add to client to get hash
		hash, err := s.TorrentClient.AddFromURLUknownScheme(t.Link, t.LinkAuth)
		if err != nil {
			log.Printf("WARNING: could not add torrent magnet for %s\n Err: %s", t.Title, err)
			s.TorrentClient.RemoveByHash(hash)
			continue
		}
		t.InfoHash = hash

		// Attempt to find a valid file
		index, terr := s.getValidFileInTorrent(t)
		if terr != nil {
			log.Printf("INFO: Passing on release %s because there was no valid file.", t.Title)
			s.TorrentClient.RemoveByHash(hash)
			continue
		}
		// TODO: Do better logic here to find "main" file
		t.MainFileIndex = index

		copyT := t // Why does this object need to be copied?? So weird..
		return &copyT, nil
	}
	return nil, nil
}

func (s *IceetimeService) getValidFileInTorrent(t Torrent) (int, error) {
	// Get correct file
	files, err := s.TorrentClient.GetFiles(t.InfoHash)
	if err != nil {
		return 0, NewError(err, 500, "unable to get files for torrent")
	}

	for i, file := range files {
		if StringEndsInAny(strings.ToLower(file), GetSupportedVideoFileFormats()) && !StringContainsAnyOf(strings.ToLower(file), GetBlacklistedFileNameContents()) {
			return i, nil
		}
	}
	return 0, NewError(nil, 400, InvalidTorrentErr)
}
