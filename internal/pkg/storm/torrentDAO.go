package storm

import (
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	"github.com/diericx/iceetime/internal/app"
	"time"
)

type TorrentDAO struct {
	db        *storm.DB
	Qualities []app.Quality
}

// NewTorrentDAO creates a new TorrentDAO object with a storm client pointed at the given path
func NewTorrentDAO(path string, qualities []app.Quality) (dao *TorrentDAO, err error) {
	db, err := storm.Open(path)
	return &TorrentDAO{
		db:        db,
		Qualities: qualities,
	}, err
}

func (dao *TorrentDAO) Close() {
	dao.db.Close()
}

// Save writes a release object to disk
func (dao *TorrentDAO) Save(t app.Torrent) error {
	t.CreatedAt = time.Now()
	return dao.db.Save(t)
}

// GetByImdbIDAndMinQuality returns a release that matches imdb id and is at least the quality specified
func (dao *TorrentDAO) GetByImdbIDAndMinQuality(imdbID string, minQuality int) (*app.Torrent, error) {
	orQualityMatchers := []q.Matcher{}
	for i := minQuality; i < len(dao.Qualities); i++ {
		quality := dao.Qualities[i]
		orQualityMatchers = append(orQualityMatchers, q.Re("TorrentName", quality.Regex))
	}

	query := dao.db.Select(
		q.Eq("ImdbID", imdbID),
		q.Or(orQualityMatchers...),
	)

	// TODO: Return multiple so we can add logic down the road if we have more than one
	// like checking the best seeders at the moment
	var torrent app.Torrent
	err := query.First(&torrent)
	if err != nil {
		// TODO: There must be a better way to check for not found...
		if err.Error() == "not found" {
			return nil, nil
		}
		return nil, err
	}

	return &torrent, err
}
