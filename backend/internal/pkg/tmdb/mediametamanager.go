package omdb

import (
	"encoding/json"
	"fmt"
	"github.com/diericx/iceetime/internal/app"
	"io/ioutil"
	"net/http"
)

const baseURL = "https://api.themoviedb.org/3/find"

// MediaMetaManager handles grabbing and returning metadata for movies using OMDB api
type MediaMetaManager struct {
	omdbApiKey string
}

// NewMediaMetaManager returns a new MediaMetaManager object
func NewMediaMetaManager(apiKey string) *MediaMetaManager {
	return &MediaMetaManager{
		omdbApiKey: apiKey,
	}
}

// GetByImdbID will query omdb for the movie and if it exists return it. Also handles caching
func (m *MediaMetaManager) GetByImdbID(imdbID string) (*app.MediaMeta, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", baseURL, imdbID), nil)
	q := req.URL.Query()
	q.Add("api_key", m.omdbApiKey)
	q.Add("external_source", "imdb_id")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Bad response from TMDB. \n Request: %s \n Status code: %v. \n Body: %s", req.URL, resp.StatusCode, string(body))
	}

	var meta app.MediaMeta

	err = json.Unmarshal(body, &meta)
	if err != nil {
		return nil, err
	}

	return &meta, nil
}
