package storm

import (
	"github.com/anacrolix/torrent/metainfo"
	"github.com/asdine/storm"
	"github.com/diericx/iceetime/internal/app"
)

type TorrentMeta struct {
	DB *storm.DB
}

func (r *TorrentMeta) Store(meta app.TorrentMeta) error {
	return r.DB.Save(&meta)
}

func (r *TorrentMeta) GetByInfoHashStr(infoHash metainfo.Hash) (app.TorrentMeta, error) {
	var meta app.TorrentMeta
	meta.InfoHash = infoHash
	err := r.DB.One("InfoHash", infoHash, &meta)
	return meta, err
}

func (r *TorrentMeta) Get() ([]app.TorrentMeta, error) {
	var metas []app.TorrentMeta
	err := r.DB.All(&metas)
	return metas, err
}

func (r *TorrentMeta) RemoveByInfoHash(infoHash metainfo.Hash) error {
	err := r.DB.DeleteStruct(app.TorrentMeta{InfoHash: infoHash})
	return err
}
