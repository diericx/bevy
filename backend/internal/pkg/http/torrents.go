package http

import (
	"net/http"

	"github.com/diericx/iceetime/internal/app"
	"github.com/gin-gonic/gin"
)

func addTorrentsGroup(rg *gin.RouterGroup, s app.TorrentService) {

	torrents := rg.Group("/torrents")
	{
		torrents.GET("", func(c *gin.Context) {
			torrents, err := s.Get()
			c.HTML(http.StatusOK, "torrents/index.tmpl", gin.H{
				"torrents": torrents,
				"error":    err,
			})
		})
	}
}
