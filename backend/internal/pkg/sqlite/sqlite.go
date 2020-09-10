package sqlite

import (
	"github.com/diericx/iceetime/internal/app"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitSqliteDB(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&app.Torrent{})

	return db, nil
}
