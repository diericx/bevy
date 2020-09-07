package http

import (
	"net/http"

	"github.com/diericx/iceetime/internal/app"
	"github.com/gin-gonic/gin"
)

func InitTorrentsGroupOnRouter(r *gin.Engine, s app.IceetimeService) {

	torrents := r.Group("/torrents")
	{
		torrents.GET("/", Index)
	}
}

func Index(c *gin.Context) {
	c.HTML(http.StatusOK, "torrents/index.tmpl", gin.H{
		"title": "Torrents",
	})
}
