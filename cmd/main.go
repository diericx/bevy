package main

import (
	"github.com/diericx/iceetime/internal/app"
	"github.com/diericx/iceetime/internal/pkg/releaseManager"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

func main() {
	releaseManagerConfig := app.ReleaseManagerConfig{}
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
	if err := d.Decode(&releaseManagerConfig); err != nil {
		log.Println("Invalid config foudn in config.yaml: ", err)
		panic(err)
	}

	log.Println(releaseManagerConfig)

	releases, _ := releases.NewReleaseManager() // TODO: handle this error

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
