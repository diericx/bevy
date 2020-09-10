package http

import (
	"github.com/diericx/iceetime/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Format struct {
	Duration int `json:"duration"`
}
type Metadata struct {
	Format Format `json:"format"`
}

type HTTPHandler struct {
	TorrentService service.TorrentService
}

func (h *HTTPHandler) Serve() {

	r := gin.Default()
	r.LoadHTMLGlob("internal/pkg/http/templates/**/*")
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "PUT", "PATCH"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	root := r.Group("/")

	addTorrentsGroup(root, h.TorrentService)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
