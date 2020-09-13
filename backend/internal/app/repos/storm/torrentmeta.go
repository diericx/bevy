package repos

import "github.com/asdine/storm"

type TorrentMeta struct {
	DB *storm.DB
}

// Store
// GetByInfoHash
