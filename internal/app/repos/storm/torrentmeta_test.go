package storm

import (
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/diericx/bevy/internal/app"

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

func TestStoreAndGetByInfoHash(t *testing.T) {
	state := setup()
	defer state.cleanup()

	meta := metainfo.MetaInfo{
		Comment: "test",
	}
	infoHash := meta.HashInfoBytes()
	expected := app.TorrentMeta{
		ID:       1,
		InfoHash: infoHash,
	}

	err := state.TorrentMetaService.Store(expected)
	if err != nil {
		t.Fatalf("Unexpected error storing: %s", err)
	}

	actual, err := state.TorrentMetaService.GetByInfoHash(infoHash)
	if err != nil {
		t.Fatalf("Unexpected error fetching: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("Expected and actual are not the same: \n EXPECTED: \n %+v \n ACTUAL: \n %+v", expected, actual)
	}
}

func TestGet(t *testing.T) {
	state := setup()
	defer state.cleanup()

	expected := []app.TorrentMeta{
		app.TorrentMeta{
			ID: 1,
			InfoHash: metainfo.MetaInfo{
				Comment: "torrent-1",
			}.HashInfoBytes(),
		},
		app.TorrentMeta{
			ID: 2,
			InfoHash: metainfo.MetaInfo{
				Comment: "torrent-2",
			}.HashInfoBytes(),
		},
	}

	for _, meta := range expected {
		err := state.TorrentMetaService.Store(meta)
		if err != nil {
			t.Fatalf("Unexpected error storing: %s", err)
		}
	}

	actual, err := state.TorrentMetaService.Get()
	if err != nil {
		t.Fatalf("Unexpected error fetching: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("Expected and actual are not the same: \n EXPECTED: \n %+v \n ACTUAL: \n %+v", expected, actual)
	}
}

func TestGetByInfoHash(t *testing.T) {
	state := setup()
	defer state.cleanup()

	meta := metainfo.MetaInfo{
		Comment: "test",
	}
	infoHash := meta.HashInfoBytes()
	expected := app.TorrentMeta{
		ID:       1,
		InfoHash: infoHash,
	}

	err := state.TorrentMetaService.Store(expected)
	if err != nil {
		t.Fatalf("Unexpected error storing: %s", err)
	}

	actual, err := state.TorrentMetaService.GetByInfoHash(infoHash)
	if err != nil {
		t.Fatalf("Unexpected error fetching: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("Expected and actual are not the same: \n EXPECTED: \n %+v \n ACTUAL: \n %+v", expected, actual)
	}
}

func Test(t *testing.T) {
	state := setup()
	defer state.cleanup()

	meta := metainfo.MetaInfo{
		Comment: "test",
	}
	infoHash := meta.HashInfoBytes()
	mockTorrentMeta := app.TorrentMeta{
		ID:       1,
		InfoHash: infoHash,
	}

	expected := []app.TorrentMeta{}

	err := state.TorrentMetaService.Store(mockTorrentMeta)
	if err != nil {
		t.Fatalf("Unexpected error storing: %s", err)
	}

	err = state.TorrentMetaService.RemoveByInfoHash(mockTorrentMeta.InfoHash)
	if err != nil {
		t.Fatalf("Unexpected error removing: %s", err)
	}

	actual, err := state.TorrentMetaService.Get()
	if err != nil {
		t.Fatalf("Unexpected error fetching: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("Expected and actual are not the same: \n EXPECTED: \n %+v \n ACTUAL: \n %+v", expected, actual)
	}
}
