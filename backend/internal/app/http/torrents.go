package http

import (
	"net/http"
	"path/filepath"

	"github.com/diericx/iceetime/internal/app"

	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type NewMagnetForm struct {
	MagnetLink string `form:"magnet_link" binding:"required"`
}

func (h *HTTPHandler) addTorrentsGroup(rg *gin.RouterGroup) {
	s := h.TorrentService

	torrents := rg.Group("/torrents")
	{
		torrents.GET("", func(c *gin.Context) {
			torrents, err := s.Get()
			c.HTML(http.StatusOK, "torrents/index.tmpl", gin.H{
				"torrents": torrents,
				"error":    err,
			})
		})

		torrents.GET("/new", func(c *gin.Context) {
			session := sessions.Default(c)

			err := session.Get("error")
			session.Set("error", nil)
			session.Save()

			c.HTML(http.StatusOK, "torrents/new.tmpl", gin.H{
				"error": err,
			})
		})

		torrents.POST("/new/magnet", func(c *gin.Context) {
			session := sessions.Default(c)

			var form NewMagnetForm
			// in this case proper binding will be automatically selected
			if err := c.ShouldBind(&form); err != nil {
				c.String(http.StatusBadRequest, "bad request")
				return
			}

			_, err := s.AddFromMagnet(form.MagnetLink, app.GetDefaultTorrentMeta())
			if err != nil {
				session.Set("error", err.Error())
				session.Save()
				c.Redirect(http.StatusFound, "/torrents/new")
				return
			}

			c.Redirect(http.StatusFound, "/torrents")
		})

		torrents.POST("/new/file", func(c *gin.Context) {
			session := sessions.Default(c)

			// Source
			file, err := c.FormFile("file")
			if err != nil {
				session.Set("error", err.Error())
				session.Save()
				c.Redirect(http.StatusFound, "/torrents/new")
				return
			}

			filename := filepath.Join(h.TorrentFilesPath, file.Filename)
			if err := c.SaveUploadedFile(file, filename); err != nil {
				session.Set("error", err.Error())
				session.Save()
				c.Redirect(http.StatusFound, "/torrents/new")
				return
			}

			_, err = s.AddFromFile(filename, app.GetDefaultTorrentMeta())
			if err != nil {
				session.Set("error", err.Error())
				session.Save()
				c.Redirect(http.StatusFound, "/torrents/new")
				return
			}

			// Remove old file now that we have one in our system
			os.Remove(filename)

			c.Redirect(http.StatusFound, "/torrents")
		})

		// torrents.GET("/stream/:infohash/file/:file", func(c *gin.Context) {
		// 	hashStr := c.Param("infohash")
		// 	fileIndexStr := c.Param("file")
		// 	fileIndex, err := strconv.ParseInt(fileIndexStr, 10, 32)
		// 	if err != nil {
		// 		c.String(http.StatusBadRequest, err.Error())
		// 	}
		// 	torrent, err := s.GetByInfoHashStr(hashStr)
		// 	if err != nil {
		// 		c.String(http.StatusBadRequest, err.Error())
		// 	}

		// 	readseeker, err := s.GetReadSeekerForFileInTorrent(torrent, int(fileIndex))
		// 	if err != nil {
		// 		c.String(http.StatusBadRequest, err.Error())
		// 	}
		// 	http.ServeContent(c.Writer, c.Request, torrent.Name, time.Time{}, readseeker)
		// })
	}
}
