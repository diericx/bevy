package http

import (
	"net/http"

	"github.com/diericx/iceetime/internal/app"
	"github.com/gin-gonic/gin"
)

func addTorrentsGroup(rg *gin.RouterGroup, s app.IceetimeService) {

	torrents := rg.Group("/torrents")
	{
		torrents.GET("/", Index)
	}
}

func Index(c *gin.Context) {
	c.HTML(http.StatusOK, "torrents/index.tmpl", gin.H{
		"title": "Torrents",
	})
}
