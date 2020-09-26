package http

import (
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/anacrolix/torrent/metainfo"

	"github.com/asdine/storm"
	"github.com/diericx/iceetime/internal/app"
	"github.com/diericx/iceetime/internal/pkg/torrent"

	"os"

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
				c.JSON(http.StatusOK, gin.H{
					"error": err.Error(),
				})
				return
			}

			filename := filepath.Join(h.TorrentFilesPath, file.Filename)
			if err := c.SaveUploadedFile(file, filename); err != nil {
				c.JSON(http.StatusOK, gin.H{
					"error": err.Error(),
				})
				return
			}

			meta := app.GetDefaultTorrentMeta()
			t, err := s.AddFromFile(filename, meta)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
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
				c.JSON(http.StatusOK, gin.H{
					"error": err.Error(),
				})
				return
			}

			meta := app.GetDefaultTorrentMeta()
			t, err := s.AddFromMagnet(input.MagnetURL, meta)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
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
					torrentResponses[i] = newTorrentResponseFromInterfaceAndMetadata(t, app.TorrentMeta{})
					continue
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
				c.JSON(http.StatusOK, gin.H{
					"error": err.Error(),
				})
				return
			}

			var infoHash metainfo.Hash
			err := infoHash.FromHexString(input.InfoHash)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"error": err.Error(),
				})
			}

			t, err := s.GetByInfoHash(infoHash)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"error": err.Error(),
				})
				return
			}

			meta, err := h.TorrentMetaRepo.GetByInfoHash(t.InfoHash())
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
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
				c.JSON(http.StatusOK, gin.H{
					"error": err.Error(),
				})
				return
			}

			var infoHash metainfo.Hash
			err := infoHash.FromHexString(input.InfoHash)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
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
				c.JSON(http.StatusOK, gin.H{
					"error": "File index not in range of files.",
				})
			}

			readseeker := files[input.FileIndex].NewReader()
			http.ServeContent(c.Writer, c.Request, t.Name(), time.Time{}, readseeker)
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
				c.JSON(http.StatusOK, gin.H{
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

			releases, err := h.ReleaseService.QueryMovie(input.ImdbID, input.Title, input.Year, input.MinQuality)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
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
					"error": err.Error(),
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
