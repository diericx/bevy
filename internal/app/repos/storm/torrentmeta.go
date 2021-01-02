package storm

import (
	"github.com/anacrolix/torrent/metainfo"
	"github.com/asdine/storm"
	"github.com/diericx/bevy/internal/app"
)

type TorrentMeta struct {
	DB *storm.DB
}

func (r *TorrentMeta) Store(meta app.TorrentMeta) error {
	return r.DB.Save(&meta)
}

// GetByInfoHash retrieves an app.TorrentMeta object from the db by info hash if exists, err if not
func (r *TorrentMeta) GetByInfoHash(infoHash metainfo.Hash) (app.TorrentMeta, error) {
	var meta app.TorrentMeta
	meta.InfoHash = infoHash
	err := r.DB.One("InfoHash", infoHash, &meta)
	return meta, err
}

// GetByTitle retrieves an app.TorrentMeta object from the db by title if exists, err if not
func (r *TorrentMeta) GetByTitle(title string) (app.TorrentMeta, error) {
	var meta app.TorrentMeta
	meta.Title = title
	err := r.DB.One("Title", title, &meta)
	return meta, err
}

// Get retrieves all TorrentMeta objects in db
func (r *TorrentMeta) Get() ([]app.TorrentMeta, error) {
	var metas []app.TorrentMeta
	err := r.DB.All(&metas)
	return metas, err
}

func (r *TorrentMeta) RemoveByInfoHash(infoHash metainfo.Hash) error {
	meta, err := r.GetByInfoHash(infoHash)
	if err != nil {
		return err
	}
	err = r.DB.DeleteStruct(&meta)
	return err
}
