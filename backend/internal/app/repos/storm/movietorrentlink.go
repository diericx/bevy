package storm

import "github.com/asdine/storm"

type MovieTorrentLink struct {
	DB *storm.DB
}

// Store
// GetByImdbID
