package http

import (
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/asdine/storm"
	"github.com/diericx/iceetime/internal/app"
	"github.com/diericx/iceetime/internal/pkg/torrent"

	"os"

	"github.com/gin-gonic/gin"
)

type NewTorrentFromMagnet struct {
	MagnetURL string `form:"magnet_url" json:"magnet_url" binding:"required"`
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

			t, err := s.AddFromFile(filename, app.GetDefaultTorrentMeta())
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
				"torrent": torrent.TorrentToStruct(t),
			})
		})

		group.POST("/torrents/new/magnet", func(c *gin.Context) {
			var json NewTorrentFromMagnet
			// in this case proper binding will be automatically selected
			if err := c.ShouldBindJSON(&json); err != nil {
				c.JSON(http.StatusOK, gin.H{
					"error": err.Error(),
				})
				return
			}

			t, err := s.AddFromMagnet(json.MagnetURL, app.GetDefaultTorrentMeta())
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"error": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"error":   nil,
				"torrent": torrent.TorrentToStruct(t),
			})
		})

		group.GET("/torrents", func(c *gin.Context) {
			torrents, err := s.Get()
			c.JSON(http.StatusOK, gin.H{
				"torrents": torrent.TorrentArrayToStructs(torrents),
				"error":    err,
			})
		})

		group.GET("/torrents/torrent/:infoHash", func(c *gin.Context) {
			infoHashStr := c.Param("infoHash")
			t, err := s.GetByInfoHashStr(infoHashStr)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"error": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"torrent": torrent.TorrentToStruct(t),
				"error":   err.Error(),
			})
		})

		group.GET("/torrents/torrent/:infoHash/stream/:file", func(c *gin.Context) {
			hashStr := c.Param("infohash")
			fileIndexStr := c.Param("file")
			fileIndex, err := strconv.ParseInt(fileIndexStr, 10, 32)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			t, err := s.GetByInfoHashStr(hashStr)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}

			readseeker, err := s.GetReadSeekerForFileInTorrent(t, int(fileIndex))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}

			http.ServeContent(c.Writer, c.Request, t.Name(), time.Time{}, readseeker)
		})

		group.GET("/torrents/find_for_movie", func(c *gin.Context) {
			imdbID := c.Query("imdb_id")
			title := c.Query("title")
			year := c.Query("year")
			minQualityStr := c.Query("min_quality")
			minQuality, err := strconv.ParseInt(minQualityStr, 10, 32)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}

			// TODO: Rename links, confusing
			links, err := h.TorrentLinkService.GetLinksForMovie(imdbID)
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
				c.JSON(http.StatusBadRequest, gin.H{
					"error":       nil,
					"torrentLink": links[0],
				})
				return
			}

			releases, err := h.ReleaseService.QueryMovie(imdbID, title, year, int(minQuality))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}

			t, fileIndex, err := h.TorrentService.AddBestTorrentFromReleases(releases, h.Qualities[minQuality])
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

			link, err := h.TorrentLinkService.LinkTorrentToMovie(imdbID, t, fileIndex)
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