package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *HTTPHandler) addTmdbGroup(group *gin.RouterGroup) {
	s := h.TmdbService

	{
		group.GET("/tmdb/browse/movies/popular", func(c *gin.Context) {
			type PopularParams struct {
				Page int `form:"page,default=1"`
			}
			var params PopularParams

			if c.Bind(&params) != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "bad query params",
				})
				return
			}

			result, err := s.PopularMovies(params.Page)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err,
				})
				return
			}
			c.JSON(http.StatusOK, result)
		})

		group.GET("/tmdb/search/movies", func(c *gin.Context) {
			type SearchParams struct {
				Query string `form:"query" binding:"required"`
				Page  int    `form:"page,default=1"`
			}
			var params SearchParams

			if c.Bind(&params) != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "bad query params",
				})
				return
			}

			result, err := s.MovieSearch(params.Query, params.Page)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err,
				})
				return
			}
			c.JSON(http.StatusOK, result)
		})

		group.GET("/tmdb/movies/:movieID", func(c *gin.Context) {
			movieIdStr := c.Params.ByName("movieID")
			movieID, err := strconv.Atoi(movieIdStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "'movieID' is not a number. Found '" + movieIdStr + "'.",
				})
				return
			}

			movie, err := s.GetMovie(movieID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err,
				})
				return
			}
			c.JSON(http.StatusOK, movie)
		})
	}
}
