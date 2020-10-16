package tmdb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type TmdbRepo struct {
	APIKey string
}

type QueryResultItem struct {
	Id           int    `json:"id"`
	Title        string `json:"title"`
	Overview     string `json:"overview"`
	PosterPath   string `json:"poster_path"`
	BackdropPath string `json:"backdrop_path"`
	ReleaseDate  string `json:"release_date"`
}

type QueryResult struct {
	Results      []QueryResultItem `json:"results"`
	TotalResults int               `json:"total_results"`
	Page         int               `json:"page"`
	TotalPages   int               `json:"total_pages"`
}

type Movie struct {
	Id           int    `json:"id"`
	Title        string `json:"title"`
	Overview     string `json:"overview"`
	PosterPath   string `json:"poster_path"`
	BackdropPath string `json:"backdrop_path"`
	ReleaseDate  string `json:"release_date"`
	ImdbId       string `json:"imdb_id"`
}

func (s *TmdbRepo) PopularMovies(page int) (QueryResult, error) {
	var result QueryResult
	url := fmt.Sprintf(
		"https://api.themoviedb.org/3/movie/popular?api_key=%s&page=%d",
		url.QueryEscape(s.APIKey),
		page,
	)

	resp, err := http.Get(url)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (s *TmdbRepo) MovieSearch(query string, page int) (QueryResult, error) {
	var result QueryResult
	url := fmt.Sprintf(
		"https://api.themoviedb.org/3/search/movie?api_key=%s&query=%s&page=%d",
		url.QueryEscape(s.APIKey),
		url.QueryEscape(query),
		page,
	)

	resp, err := http.Get(url)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (s *TmdbRepo) GetMovie(id int) (Movie, error) {
	var result Movie
	url := fmt.Sprintf(
		"https://api.themoviedb.org/3/movie/%d?api_key=%s",
		id,
		url.QueryEscape(s.APIKey),
	)

	resp, err := http.Get(url)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}
