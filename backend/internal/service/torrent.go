package service

import (
	"errors"
	"log"

	"github.com/diericx/iceetime/internal/app"
)

type Torrent struct {
	app.Torrent
	app.TorrentStats
	PercentageCompleted int
}

type TorrentService struct {
	TorrentDAO    app.TorrentDAO
	TorrentClient app.TorrentClient
}

func (s *TorrentService) Add(torrent app.Torrent) (Torrent, error) {
	var err error
	if torrent.MagnetLink != "" {
		torrent, err = s.TorrentClient.AddFromMagnet(torrent)
		if err != nil {
			return Torrent{}, err
		}
	} else if torrent.File != "" {
		torrent, err = s.TorrentClient.AddFromFile(torrent)
		if err != nil {
			return Torrent{}, err
		}
	} else {
		return Torrent{}, errors.New("must specify magnet or file")
	}

	// TODO: should we actually start downloading? Will this mess up readseekers?
	s.TorrentClient.Start(torrent)

	// Save to db and return copy with updated fields from DB
	torrent, err = s.TorrentDAO.Store(torrent)
	if err != nil {
		return Torrent{}, err
	}

	// Return (without getting stats)
	return Torrent{
		Torrent: torrent,
	}, nil
}

func (s *TorrentService) GetByID(id uint) (app.Torrent, error) {
	// TODO: fill in status from torrent client
	return s.TorrentDAO.GetByID(id)
}

func (s *TorrentService) Get() ([]Torrent, error) {
	torrents, err := s.TorrentDAO.Get()
	if err != nil {
		return nil, err
	}

	torrentsToReturn := make([]Torrent, len(torrents))
	for i, torrent := range torrents {
		stats, err := s.TorrentClient.Stats(torrent)
		if err != nil {
			log.Println(err)
			log.Printf("%+v", torrent)
			torrentsToReturn[i] = Torrent{
				Torrent: torrent,
			}
		}

		torrentsToReturn[i] = Torrent{
			Torrent:             torrent,
			TorrentStats:        stats,
			PercentageCompleted: int(100 * (stats.BytesCompleted + 1) / (torrent.Size + 1)),
		}
	}
	// TODO: fill in status from torrent client
	return torrentsToReturn, nil
}
