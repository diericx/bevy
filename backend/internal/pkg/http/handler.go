package http

import (
	"github.com/diericx/iceetime/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

type Format struct {
	Duration int `json:"duration"`
}
type Metadata struct {
	Format Format `json:"format"`
}

type HTTPHandler struct {
	TorrentService  service.TorrentService
	TorrentFilePath string
}

func (h *HTTPHandler) Serve(cookieSecret string) {

	r := gin.Default()
	store := cookie.NewStore([]byte(cookieSecret))
	r.Use(sessions.Sessions("mysession", store))

	r.LoadHTMLGlob("internal/pkg/http/templates/**/*")
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "PUT", "PATCH"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	root := r.Group("/")

	h.addTorrentsGroup(root)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
