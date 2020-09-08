package http

import (
	"net/http"

	"github.com/diericx/iceetime/internal/app"
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
	IceetimeService app.IceetimeService
}

func (h *HTTPHandler) Serve() {
	iceetimeService := h.IceetimeService

	r := gin.Default()
	r.LoadHTMLGlob("internal/pkg/http/templates/**/*")

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "PUT", "PATCH"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	v1 := r.Group("/v1")

	addTorrentsGroup(v1, iceetimeService)
	addStreamRoutes(v1, iceetimeService)

	r.GET("/find/movie", func(c *gin.Context) {
		imdbID := c.Query("imdbid")
		title := c.Query("title")
		year := c.Query("year")
		if imdbID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   true,
				"message": "Missing imdb id",
			})
			return
		}
		if title == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   true,
				"message": "Missing title",
			})
			return
		}
		if year == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   true,
				"message": "Missing year",
			})
			return
		}

		torrent, err := iceetimeService.FindLocallyOrFetchMovie(imdbID, title, year, 1)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   true,
				"message": err.Message,
			})
			return
		}

		c.JSON(200, torrent)
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
