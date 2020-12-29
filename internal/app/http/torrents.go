package http

import (
	"errors"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/anacrolix/torrent/metainfo"

	"github.com/asdine/storm"
	"github.com/diericx/iceetime/internal/app"
	"github.com/diericx/iceetime/internal/pkg/torrent"

	"os"

	"github.com/jinzhu/copier"

	"github.com/gin-gonic/gin"
)

type Torrent struct {
	app.TorrentMeta
	// Comes from torrent interface
	BytesCompleted int64  `json:"bytesCompleted"`
	Length         int64  `json:"length"`
	InfoHash       string `json:"infoHash"`
	Name           string `json:"name"`
	TotalPeers     int    `json:"totalPeers"`
	ActivePeers    int    `json:"activePeers"`
}

type Release struct {
	ImdbID       string  `json:"imdbId"`
	Title        string  `json:"title"`
	Size         int64   `json:"size"`
	InfoHash     string  `json:"infoHash"`
	Grabs        int     `json:"grabs"`
	Seeders      int     `json:"seeders"`
	MinRatio     float32 `json:"minRatio"`
	MinSeedTime  int     `json:"minSeedTime"`
	SeederScore  float64 `json:"seederScore"`
	SizeScore    float64 `json:"sizeScore"`
	QualityScore float64 `json:"qualityScore"`
	AlreadyAdded bool    `json:"alreadyAdded"`
}

func newTorrentResponseFromInterfaceAndMetadata(tIn torrent.Torrent, meta app.TorrentMeta) Torrent {
	return Torrent{
		TorrentMeta:    meta,
		BytesCompleted: tIn.BytesCompleted(),
		Length:         tIn.Length(),
		InfoHash:       tIn.InfoHash().HexString(),
		Name:           tIn.Name(),
		TotalPeers:     tIn.Stats().TotalPeers,
		ActivePeers:    tIn.Stats().ActivePeers,
	}
}

func (h *HTTPHandler) addTorrentsGroup(group *gin.RouterGroup) {
	s := h.TorrentService

	{
		group.POST("/torrents/new/file", func(c *gin.Context) {
			// Source
			file, err := c.FormFile("file")
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}

			filename := filepath.Join(h.TorrentFilesPath, file.Filename)
			if err := c.SaveUploadedFile(file, filename); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}

			meta := app.GetDefaultTorrentMeta()
			t, err := s.AddFromFile(filename, meta)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}

			// Remove old file now that we have one in our system
			os.Remove(filename)

			c.JSON(http.StatusOK, gin.H{
				"error":   nil,
				"torrent": newTorrentResponseFromInterfaceAndMetadata(t, meta),
			})
		})

		group.POST("/torrents/new/magnet", func(c *gin.Context) {
			type Input struct {
				MagnetURL string `form:"magnet_url" json:"magnet_url" binding:"required"`
			}

			var input Input
			// in this case proper binding will be automatically selected
			if err := c.ShouldBindJSON(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}

			meta := app.GetDefaultTorrentMeta()
			t, err := s.AddFromMagnet(input.MagnetURL, meta)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"error":   nil,
				"torrent": newTorrentResponseFromInterfaceAndMetadata(t, meta),
			})
		})

		group.GET("/torrents", func(c *gin.Context) {
			torrents := s.Get()
			torrentResponses := make([]Torrent, len(torrents))

			for i, t := range torrents {
				meta, err := h.TorrentMetaRepo.GetByInfoHash(t.InfoHash())
				if err != nil {
					c.JSON(http.StatusOK, gin.H{
						"error": err,
					})
					return
				}

				torrentResponses[i] = newTorrentResponseFromInterfaceAndMetadata(t, meta)
			}

			c.JSON(http.StatusOK, gin.H{
				"torrents": torrentResponses,
				"error":    false,
			})
		})

		group.GET("/torrents/torrent/:infoHash", func(c *gin.Context) {
			type Input struct {
				InfoHash string `uri:"infoHash" binding:"required"`
			}
			var input Input

			if err := c.ShouldBindUri(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}

			var infoHash metainfo.Hash
			err := infoHash.FromHexString(input.InfoHash)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
			}

			t, err := s.GetByInfoHash(infoHash)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}

			meta, err := h.TorrentMetaRepo.GetByInfoHash(t.InfoHash())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
			}

			c.JSON(http.StatusOK, gin.H{
				"torrent": newTorrentResponseFromInterfaceAndMetadata(t, meta),
				"error":   err.Error(),
			})
		})

		group.GET("/torrents/torrent/:infoHash/stream/:file", func(c *gin.Context) {
			type Input struct {
				InfoHash  string `uri:"infoHash" binding:"required"`
				FileIndex uint   `uri:"file"`
			}
			var input Input

			if err := c.ShouldBindUri(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}

			var infoHash metainfo.Hash
			err := infoHash.FromHexString(input.InfoHash)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
			}

			t, err := s.GetByInfoHash(infoHash)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}

			files := t.Files()
			if int(input.FileIndex) > len(files) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "File index not in range of files.",
				})
			}

			readseeker := files[input.FileIndex].NewReader()
			http.ServeContent(c.Writer, c.Request, t.Name(), time.Time{}, readseeker)
		})

		group.GET("/torrents/releases_for_movie", func(c *gin.Context) {
			type Input struct {
				ImdbID     string `form:"imdb_id" binding:"required"`
				Title      string `form:"title" binding:"required"`
				Year       string `form:"year" binding:"required"`
				MinQuality int    `form:"min_quality"`
			}
			var input Input

			if err := c.ShouldBind(&input); err != nil {
				log.Println("Error binding json: ", err)
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}

			scoredReleases, err := h.ReleaseService.QueryMovie(input.ImdbID, input.Title, input.Year)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}

			// convert releases to release response
			releaseResponses := make([]Release, len(scoredReleases))
			for i, scoredRelease := range scoredReleases {
				releaseResponse := Release{}
				copier.Copy(&releaseResponse, &scoredRelease)

				_, err := s.GetByTitle(scoredRelease.Title)
				if err != nil {
					releaseResponse.AlreadyAdded = false
				} else {
					releaseResponse.AlreadyAdded = true
				}

				releaseResponses[i] = releaseResponse
			}

			c.JSON(http.StatusOK, gin.H{
				"error":    nil,
				"releases": releaseResponses,
			})
		})

		group.GET("/torrents/find_for_movie", func(c *gin.Context) {
			type Input struct {
				ImdbID     string `form:"imdb_id" binding:"required"`
				Title      string `form:"title" binding:"required"`
				Year       string `form:"year" binding:"required"`
				MinQuality int    `form:"min_quality"`
			}
			var input Input

			if err := c.ShouldBind(&input); err != nil {
				log.Println("Error binding json: ", err)
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}

			// TODO: Rename links, confusing
			links, err := h.TorrentLinkService.GetLinksForMovie(input.ImdbID)
			if err != nil {
				if err != storm.ErrNotFound {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": err.Error(),
					})
					return
				}
			}
			// Return the torrent if one is already linked to this movie
			if len(links) > 0 {
				c.JSON(http.StatusOK, gin.H{
					"error":       nil,
					"torrentLink": links[0],
				})
				return
			}

			scoredReleases, err := h.ReleaseService.QueryMovie(input.ImdbID, input.Title, input.Year)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}

			// Convert scored releases back to regular releases
			releases := make([]app.Release, len(scoredReleases))
			for i, scoredRelease := range scoredReleases {
				releases[i] = scoredRelease.Release
			}

			t, fileIndex, err := h.TorrentService.AddBestTorrentFromReleases(releases, h.Qualities[input.MinQuality])
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			if t == nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": errors.New("no torrent was found"),
				})
				return
			}

			link, err := h.TorrentLinkService.LinkTorrentToMovie(input.ImdbID, t, fileIndex)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"error":       nil,
				"torrentLink": link,
			})
		})
	}
}
