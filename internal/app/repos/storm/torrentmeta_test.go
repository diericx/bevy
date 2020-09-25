package storm

import (
	"log"
	"os"
	"testing"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/diericx/iceetime/internal/app"

	"github.com/asdine/storm"
)

type state struct {
	DB                 *storm.DB
	TorrentMetaService TorrentMeta
}

func setup() state {
	stormDBFilePath := ".test.storm.db"
	stormDB, err := OpenDB(stormDBFilePath)
	if err != nil {
		log.Fatalf("Couldn't open torrent file at %s. The file will be created if it doesn't exist, make sure the directory exists and user has proper permissions.", stormDBFilePath)
	}

	return state{
		TorrentMetaService: TorrentMeta{
			DB: stormDB,
		},
	}
}

func (s state) cleanup() {
	s.TorrentMetaService.DB.Close()
	os.Remove(".test.storm.db")
}

func TestStore(t *testing.T) {
	state := setup()
	defer state.cleanup()

	meta := metainfo.MetaInfo{
		Comment: "test",
	}
	infoHash := meta.HashInfoBytes()
	expectedMeta := app.TorrentMeta{
		InfoHash: infoHash,
	}

	err := state.TorrentMetaService.Store(expectedMeta)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

}
