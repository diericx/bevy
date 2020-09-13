package storm

import (
	"github.com/asdine/storm"
	"github.com/diericx/iceetime/internal/app"
)

type TorrentMeta struct {
	DB *storm.DB
}

func (r *TorrentMeta) Store(meta app.TorrentMeta) error {
	return r.DB.Save(&meta)
}
func (r *TorrentMeta) GetByInfoHash(infoHashStr string) (app.TorrentMeta, error) {
	var meta app.TorrentMeta
	err := r.DB.One("InfoHash", infoHashStr, &meta)
	return meta, err
}

// Store
// GetByInfoHash
