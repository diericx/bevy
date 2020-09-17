package jackett

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/diericx/iceetime/internal/app"
)

type Rss struct {
	XMLName  xml.Name `xml:"rss"`
	indexer  app.Indexer
	Channels []Channel `xml:"channel"`
}

type Channel struct {
	XMLName xml.Name `xml:"channel"`
	Title   string   `xml:"title"`
	Items   []Item   `xml:"item"`
}

type TorznabAttr struct {
	Key   string `xml:"name,attr"`
	Value int    `xml:"value,attr"`
}

type Item struct {
	XMLName        xml.Name               `xml:"item"`
	Title          string                 `xml:"title"`
	Size           int64                  `xml:"size"`
	Grabs          int                    `xml:"grabs"`
	Link           string                 `xml:"link"`
	TorznabAttrMap map[string]interface{} // Assumption: all torznab attributes are numbers
	TorznabAttrs   []TorznabAttr          `xml:"attr"` // Used for immediate parsing but map is more useful
}

type ReleaseRepo struct {
	Qualities []app.Quality
	Indexers  []app.Indexer
}

// NewReleaseRepo instantiates a new ReleaseRepo object that implements torznab queries/indexers
func NewReleaseRepo(indexers []app.Indexer, qualities []app.Quality) (*ReleaseRepo, error) {
	return &ReleaseRepo{
		Indexers:  indexers,
		Qualities: qualities,
	}, nil
}

func (s *ReleaseRepo) QueryAllIndexers(imdbID string, queryString string) ([]app.Release, error) {
	torznabResponses, err := s.torznabQuery(imdbID, queryString)
	if err != nil {
		return nil, err
	}

	releases := []app.Release{}
	for _, resp := range torznabResponses {
		for _, channel := range resp.Channels {
			for _, item := range channel.Items {
				releases = append(releases, app.Release{
					ImdbID:   imdbID,
					Title:    item.Title,
					Size:     item.Size,
					Link:     item.Link,
					LinkAuth: resp.indexer.BasicAuth,
					// TODO: Handle assertion errors
					InfoHash:    getStringFromMap(item.TorznabAttrMap, "infohash", ""),
					Grabs:       getIntFromMap(item.TorznabAttrMap, "grabs", 0),
					Seeders:     getIntFromMap(item.TorznabAttrMap, "seeders", 0),
					MinRatio:    getFloat32FromMap(item.TorznabAttrMap, "minratio", 0),
					MinSeedTime: getIntFromMap(item.TorznabAttrMap, "minseedtime", 0),
				})
			}
		}
	}

	return releases, nil
}

func (s *ReleaseRepo) torznabQuery(imdbID string, search string) ([]Rss, error) {
	torznabResponses := []Rss{}

	for _, indexer := range s.Indexers {
		var torznabResp Rss
		torznabResp.indexer = indexer

		client := &http.Client{}

		req, err := http.NewRequest("GET", indexer.URL, nil)
		if indexer.BasicAuth != nil {
			req.SetBasicAuth(indexer.BasicAuth.Username, indexer.BasicAuth.Password)
		}
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Add("apikey", indexer.APIKey)

		if indexer.SupportsImdbIDSearch {
			q.Add("t", "movie")
			q.Add("imdbid", imdbID)
		} else {
			q.Add("t", "search")
			q.Add("q", search)
		}

		q.Add("cat", indexer.Categories) //TODO: Make a difference between tv and movies
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
			return nil, fmt.Errorf("%v: %s", resp.StatusCode, string(body))
		}

		xml.Unmarshal(body, &torznabResp)

		// Convert torznab attr array to map
		for _, channel := range torznabResp.Channels {
			for i, item := range channel.Items {
				if item.TorznabAttrMap == nil {
					item.TorznabAttrMap = make(map[string]interface{})
				}
				for _, attr := range item.TorznabAttrs {
					item.TorznabAttrMap[attr.Key] = attr.Value
				}
				channel.Items[i] = item
			}
		}
		torznabResponses = append(torznabResponses, torznabResp)
	}
	return torznabResponses, nil
}

func getStringFromMap(m map[string]interface{}, key string, d string) string {
	v, ok := m[key].(string)
	if !ok {
		return d
	}
	return v
}

func getIntFromMap(m map[string]interface{}, key string, d int) int {
	v, ok := m[key].(int)
	if !ok {
		return d
	}
	return v
}

func getFloat32FromMap(m map[string]interface{}, key string, d float32) float32 {
	v, ok := m[key].(float32)
	if !ok {
		return d
	}
	return v
}
