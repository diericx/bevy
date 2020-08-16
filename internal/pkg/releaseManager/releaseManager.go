package releases

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

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
	XMLName        xml.Name       `xml:"item"`
	Title          string         `xml:"title"`
	Size           int64          `xml:"size"`
	Grabs          int            `xml:"grabs"`
	Link           string         `xml:"link"`
	TorznabAttrMap map[string]int // Assumption: all torznab attributes are numbers
	TorznabAttrs   []TorznabAttr  `xml:"attr"` // Used for immediate parsing but map is more useful
}

type rmanager struct {
	Qualities []app.Quality
	Indexers  []app.Indexer
}

func NewReleaseManager(indexers []app.Indexer, qualities []app.Quality) (*rmanager, *app.Error) {
	return &rmanager{
		Indexers:  indexers,
		Qualities: qualities,
	}, nil
}

func (r *rmanager) GetByImdbIDAndQuality(imdbID string, minQuality int) (*app.Release, *app.Error) {
	// TODO: Return release saved locally or null
	return &app.Release{}, nil
}

func (r *rmanager) AddFromTorznabQuery(imdbID string, minQuality int) (*app.Release, *app.Error) {
	releases, err := r.torznabQuery("movie", imdbID, r.Qualities[minQuality].Regex, "2040")
	if err != nil {
		return nil, app.NewError(err, 500, "Error querying torznab")
	}
	log.Printf("%+v", releases)
	// TODO: Query Jackett for releases
	// TODO: Find best release
	// TODO: Save and return release
	return &app.Release{}, nil
}

func (r *rmanager) torznabQuery(t string, imdbID string, search string, categories string) (*Rss, *app.Error) {
	var torznabResp Rss

	for _, indexer := range r.Indexers {
		client := &http.Client{}

		req, err := http.NewRequest("GET", indexer.URL, nil)
		if indexer.BasicAuth != nil {
			req.SetBasicAuth(indexer.BasicAuth.Username, indexer.BasicAuth.Password)
		}
		q := req.URL.Query()
		q.Add("apikey", indexer.APIKey)
		q.Add("t", t)
		q.Add("imdbid", imdbID)
		q.Add("cat", categories)
		req.URL.RawQuery = q.Encode()

		resp, err := client.Do(req)
		if err != nil {
			return nil, app.NewError(err, 500, fmt.Sprintf("Error querying indexer %s", indexer.Name))
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, app.NewError(err, 500, fmt.Sprintf("Error parsing response body from indexer %s", indexer.Name))
		}

		xml.Unmarshal(body, &torznabResp)

		// Convert torznab attr array to map
		for _, channel := range torznabResp.Channels {
			for i, item := range channel.Items {
				if item.TorznabAttrMap == nil {
					item.TorznabAttrMap = make(map[string]int)
				}
				for _, attr := range item.TorznabAttrs {
					item.TorznabAttrMap[attr.Key] = attr.Value
				}
				channel.Items[i] = item
			}
		}
	}
	return &torznabResp, nil
}
