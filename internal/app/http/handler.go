package http

import (
	"html/template"

	"github.com/diericx/iceetime/internal/app"
	"github.com/diericx/iceetime/internal/app/services"
	"github.com/diericx/iceetime/internal/pkg/torrent"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

type Format struct {
	Duration int `json:"duration"`
}
type Metadata struct {
	Format Format `json:"format"`
}

type HTTPHandler struct {
	TorrentService     services.Torrent
	ReleaseService     services.Release
	TorrentLinkService services.TorrentLink
	Transcoder         services.Transcoder
	Qualities          []app.Quality
	TorrentFilesPath   string
}

func (h *HTTPHandler) Serve(cookieSecret string) {

	r := gin.Default()
	store := cookie.NewStore([]byte(cookieSecret))
	r.Use(sessions.Sessions("mysession", store))

	r.SetFuncMap(template.FuncMap{
		"getTorrentProg": getTorrentProg,
	})

	// Serve frontend static files
	r.Use(static.Serve("/", static.LocalFile("./frontend/build", true)))

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "PUT", "PATCH"},
		AllowHeaders:     []string{"Origin, Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	v1 := r.Group("/v1")

	h.addTorrentsGroup(v1)
	h.addTranscoderGroup(v1)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func getTorrentProg(t torrent.Torrent) int64 {
	if t.Length() == 0 {
		return 0
	}
	return 100 * t.BytesCompleted() / t.Length()
}
