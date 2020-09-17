package storm

import (
	"github.com/asdine/storm"
)

func OpenDB(path string) (*storm.DB, error) {
	return storm.Open(path)
}
