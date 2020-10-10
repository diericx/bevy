package services

import (
	"strings"

	"github.com/diericx/iceetime/internal/app/repos/tmdb"
)

type Tmdb struct {
	TmdbRepo tmdb.TmdbRepo
}

type QueryResultItem struct {
	Id          int    `json:"id"`
	Title       string `json:"title"`
	Overview    string `json:"overview"`
	PosterUrl   string `json:"poster_url"`
	BackdropUrl string `json:"backdrop_url"`
	ReleaseYear string `json:"release_year"`
}

type QueryResult struct {
	Results      []QueryResultItem `json:"results"`
	TotalResults int               `json:"total_results"`
	Page         int               `json:"page"`
	TotalPages   int               `json:"total_pages"`
}

type Movie struct {
	Id          int    `json:"id"`
	Title       string `json:"title"`
	Overview    string `json:"overview"`
	PosterUrl   string `json:"poster_url"`
	BackdropUrl string `json:"backdrop_url"`
	ReleaseYear string `json:"release_year"`
	ImdbId      string `json:"imdb_id"`
}

func (s *Tmdb) PopularMovies(page int) (QueryResult, error) {
	var result QueryResult

	repoResult, err := s.TmdbRepo.PopularMovies(page)
	if err != nil {
		return result, err
	}

	result, err = ConvertQueryResult(repoResult)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (s *Tmdb) MovieSearch(query string, page int) (QueryResult, error) {
	var result QueryResult

	repoResult, err := s.TmdbRepo.MovieSearch(query, page)
	if err != nil {
		return result, err
	}

	result, err = ConvertQueryResult(repoResult)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (s *Tmdb) GetMovie(id int) (Movie, error) {
	var movie Movie

	repoMovie, err := s.TmdbRepo.GetMovie(id)
	if err != nil {
		return movie, err
	}

	movie, err = ConvertMovie(repoMovie)
	if err != nil {
		return movie, err
	}
	return movie, nil
}

func ConvertQueryResult(repoResult tmdb.QueryResult) (QueryResult, error) {
	var result QueryResult

	result.Page = repoResult.Page
	result.TotalPages = repoResult.TotalPages
	result.TotalResults = repoResult.TotalResults
	result.Results = make([]QueryResultItem, len(repoResult.Results))

	for i, s := range repoResult.Results {

		result.Results[i].Id = s.Id
		result.Results[i].Title = s.Title
		result.Results[i].Overview = s.Overview

		// convert poster/backdrop paths to urls
		result.Results[i].PosterUrl = posterPathToPosterUrl(s.PosterPath)
		result.Results[i].BackdropUrl = posterPathToPosterUrl(s.BackdropPath)

		// convert dates to years
		year, err := releaseDateToReleaseYear(s.ReleaseDate)
		if err != nil {
			return result, err
		}
		result.Results[i].ReleaseYear = year
	}

	return result, nil
}

func ConvertMovie(repoMovie tmdb.Movie) (Movie, error) {
	var movie Movie

	movie.Id = repoMovie.Id
	movie.Title = repoMovie.Title
	movie.Overview = repoMovie.Overview
	movie.ImdbId = repoMovie.ImdbId

	// convert poster/backdrop paths to urls
	movie.PosterUrl = posterPathToPosterUrl(repoMovie.PosterPath)
	movie.BackdropUrl = backdropPathToBackdropUrl(repoMovie.BackdropPath)

	// convert dates to years
	year, err := releaseDateToReleaseYear(repoMovie.ReleaseDate)
	if err != nil {
		return movie, err
	}
	movie.ReleaseYear = year

	return movie, nil
}

func releaseDateToReleaseYear(releaseDate string) (string, error) {
	if releaseDate == "" {
		return "", nil
	}
	dateSplit := strings.Split(releaseDate, "-")
	return dateSplit[0], nil
}

func posterPathToPosterUrl(posterPath string) string {
	if posterPath != "" {
		return "https://image.tmdb.org/t/p/w500" + posterPath
	}
	return ""
}

func backdropPathToBackdropUrl(backdropPath string) string {
	if backdropPath != "" {
		return "https://image.tmdb.org/t/p/original" + backdropPath
	}
	return ""
}
