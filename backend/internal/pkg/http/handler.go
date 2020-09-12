package http

import (
	"html/template"

	"github.com/diericx/iceetime/internal/app"
	"github.com/diericx/iceetime/internal/pkg/ffmpeg"
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
	TorrentService  app.TorrentService
	TorrentFilePath string
	Transcoder      ffmpeg.Transcoder
}

func (h *HTTPHandler) Serve(cookieSecret string) {

	r := gin.Default()
	store := cookie.NewStore([]byte(cookieSecret))
	r.Use(sessions.Sessions("mysession", store))

	r.SetFuncMap(template.FuncMap{
		"getTorrentProg": getTorrentProg,
	})

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
	h.addTranscoderGroup(root)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func getTorrentProg(t app.Torrent) int64 {
	if t.Length == 0 {
		return 0
	}
	return 100 * t.BytesCompleted / t.Length
}
