package releases

import "github.com/diericx/iceetime/internal/app"

type rmanager struct {
}

func NewReleaseManager() (*rmanager, *app.Error) {
	return &rmanager{}, nil
}

func (r *rmanager) Get(imdbID string, minQuality app.Quality) (*app.Release, *app.Error) {
	// TODO: Return release saved locally or null
	return &app.Release{}, nil
}

func (r *rmanager) Add(imdbID string, minQuality app.Quality) (*app.Release, *app.Error) {
	// TODO: Query Jackett for releases
	// TODO: Find best release
	// TODO: Save and return release
	return &app.Release{}, nil
}

// func (r *rmanager) fetchAllExistingReleases(search string, categories []string) ([]app.Release, *app.Error) {

// }
