package http

import (
	"net/http"

	"github.com/diericx/iceetime/internal/app"
	"github.com/diericx/iceetime/internal/service"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type NewForm struct {
	MagnetLink string `form:"magnet_link" binding:"required"`
}

func addTorrentsGroup(rg *gin.RouterGroup, s service.TorrentService) {

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

		torrents.POST("/new", func(c *gin.Context) {
			session := sessions.Default(c)
			var form NewForm
			// in this case proper binding will be automatically selected
			if err := c.ShouldBind(&form); err != nil {
				c.String(http.StatusBadRequest, "bad request")
				return
			}

			_, err := s.Add(app.Torrent{MagnetLink: form.MagnetLink})
			if err != nil {
				session.Set("error", err.Error())
				session.Save()
				c.Redirect(http.StatusFound, "/torrents/new")
				return
			}

			c.Redirect(http.StatusFound, "/torrents")
		})
	}
}
