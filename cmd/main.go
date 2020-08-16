package main

import (
	"log"
	"os"

	"github.com/diericx/iceetime/internal/app"
	releases "github.com/diericx/iceetime/internal/pkg/releaseManager"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

func main() {
	config := app.Config{}

	db, err := storm.Open("iceetime.db")
	defer db.Close()

	// Open release manager config file
	file, err := os.Open("./config.yaml")
	if err != nil {
		log.Println("Config file not found: config.yaml")
		return
	}
	defer file.Close()
	// Init new YAML decode
	d := yaml.NewDecoder(file)
	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		log.Println("Invalid config foudn in config.yaml: ", err)
		panic(err)
	}

	log.Println(config.Indexers)

	releases, _ := releases.NewReleaseManager(config.Indexers) // TODO: handle this error

	r := gin.Default()
	r.GET("/find/movie/imdb_id/:imdbID", func(c *gin.Context) {
		imdbID := c.Param("imdbID")
		foundReleases, _ := releases.Get(imdbID, app.Quality{}) // TODO: manage this error
		c.JSON(200, gin.H{
			"releases": foundReleases,
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
