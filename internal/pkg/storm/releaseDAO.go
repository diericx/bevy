package storm

import (
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	"github.com/diericx/iceetime/internal/app"
)

type ReleaseDAO struct {
	db        *storm.DB
	Qualities []app.Quality
}

// NewReleaseDAO creates a new ReleaseDAO object with a storm client pointed at the given path
func NewReleaseDAO(path string) (dao *ReleaseDAO, err error) {
	db, err := storm.Open(path)
	return &ReleaseDAO{
		db: db,
	}, err
}

// Save writes a release object to disk
func (dao *ReleaseDAO) Save(Release app.Release) error {
	return dao.db.Save(Release)
}

// GetByImdbIDAndMinQuality returns a release that matches imdb id and is at least the quality specified
func (dao *ReleaseDAO) GetByImdbIDAndMinQuality(imdbID string, minQuality int) (app.Release, error) {
	orQualityMatchers := []q.Matcher{}
	for i := minQuality; i < len(dao.Qualities); i++ {
		quality := dao.Qualities[i]
		orQualityMatchers = append(orQualityMatchers, q.Re("TorrentName", quality.Regex))
	}

	var release app.Release
	query := dao.db.Select(
		q.Eq("ImdbID", imdbID),
		q.Or(orQualityMatchers...),
	)
	// TODO: Return multiple so we can add logic if we have more than one
	err := query.First(release)
	return release, err
}
