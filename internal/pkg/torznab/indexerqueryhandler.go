package torznab

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/diericx/iceetime/internal/app"
)

type Rss struct {
	XMLName  xml.Name  `xml:"rss"`
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

type indexerQueryHandler struct {
	MediaMetaManager app.MediaMetaManager
	Qualities        []app.Quality
	Indexers         []app.Indexer
}

// NewIndexerQueryHandler instantiates a new IndexerQueryHandler object that implements torznab queries/indexers
func NewIndexerQueryHandler(mediaMetaManager app.MediaMetaManager, indexers []app.Indexer, qualities []app.Quality) (*indexerQueryHandler, *app.Error) {
	return &indexerQueryHandler{
		MediaMetaManager: mediaMetaManager,
		Indexers:         indexers,
		Qualities:        qualities,
	}, nil
}

func (iqh *indexerQueryHandler) QueryMovie(imdbID string, title string, year string, minQuality int) (*app.Torrent, *app.Error) {
	quality := iqh.Qualities[minQuality] // TODO: Go through each quality attempting to get releases instead of just the min

	torznabResponses, err := iqh.torznabQuery(imdbID, fmt.Sprintf("%s %s %s", title, year, iqh.Qualities[minQuality].Regex))
	if err != nil {
		return nil, err
	}

	bestScore := 0.0
	var bestRelease Item
	for _, resp := range torznabResponses {
		for _, channel := range resp.Channels {
			for _, item := range channel.Items {
				score := 0.0
				if item.Size < quality.MinSize || item.Size > quality.MaxSize {
					log.Printf("Passing on release %s because size %v is not correct.", item.Title, item.Size)
					continue
				}

				// TODO: Can we just cast to float? probably...
				score += float64(item.TorznabAttrMap["seeders"].(int)) / 10

				if score > bestScore {
					bestScore = score
					bestRelease = item
				}
			}
		}
	}

	torrent := app.Torrent{
		ImdbID:   imdbID,
		Title:    bestRelease.Title,
		Size:     bestRelease.Size,
		FileLink: bestRelease.Link,
		// TODO: Handle assertion errors
		InfoHash:    getStringFromMap(bestRelease.TorznabAttrMap, "infohash", ""),
		Grabs:       getIntFromMap(bestRelease.TorznabAttrMap, "grabs", 0),
		Seeders:     getIntFromMap(bestRelease.TorznabAttrMap, "seeders", 0),
		MinRatio:    getFloat32FromMap(bestRelease.TorznabAttrMap, "minratio", 0),
		MinSeedTime: getIntFromMap(bestRelease.TorznabAttrMap, "minseedtime", 0),
	}

	if strings.Contains(bestRelease.Link, "magnet:?") {
		torrent.MagnetLink = bestRelease.Link
	} else {
		torrent.FileLink = bestRelease.Link
	}

	log.Printf("Best: %+v, %+v", bestScore, bestRelease)

	return &torrent, nil
}

func (iqh *indexerQueryHandler) torznabQuery(imdbID string, search string) ([]Rss, *app.Error) {
	torznabResponses := []Rss{}

	for _, indexer := range iqh.Indexers {
		var torznabResp Rss
		client := &http.Client{}

		req, err := http.NewRequest("GET", indexer.URL, nil)
		if indexer.BasicAuth != nil {
			req.SetBasicAuth(indexer.BasicAuth.Username, indexer.BasicAuth.Password)
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
		log.Printf("CURL: %s", req.URL)

		resp, err := client.Do(req)
		if err != nil {
			return nil, app.NewError(err, 500, fmt.Sprintf("Error querying indexer %s", indexer.Name))
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, app.NewError(err, 500, fmt.Sprintf("Error parsing response body from indexer %s", indexer.Name))
		}

		// log.Println(string(body))
		if resp.StatusCode != 200 {
			return nil, app.NewError(nil, 500, fmt.Sprintf("Error making query to indexer. Body of request: %s", string(body)))
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
