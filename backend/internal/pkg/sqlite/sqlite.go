package sqlite

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitSqliteDB(path string) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(path), &gorm.Config{})
}
