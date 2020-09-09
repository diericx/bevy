package sqlite

import (
	"github.com/diericx/iceetime/internal/app"
	"gorm.io/gorm"
)

type TorrentDAO struct {
	Db *gorm.DB
}

func (dao *TorrentDAO) Store(torrent app.Torrent) (app.Torrent, error) {
	result := dao.Db.Create(&torrent)
	return torrent, result.Error
}

func (dao *TorrentDAO) GetByID(id uint) (app.Torrent, error) {
	var torrent app.Torrent
	result := dao.Db.First(&torrent, id)
	return torrent, result.Error
}

func (dao *TorrentDAO) Get() ([]app.Torrent, error) {
	var torrents []app.Torrent
	result := dao.Db.Find(&torrents)
	return torrents, result.Error
}

func (dao *TorrentDAO) Remove(id uint) error {
	result := dao.Db.Delete(&app.Torrent{}, id)
	return result.Error
}
