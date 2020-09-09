package sqlite

import (
	"os"
	"testing"

	"github.com/diericx/iceetime/internal/app"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setup(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to db: %s", err)
	}
	db.AutoMigrate(&app.Torrent{})
	return db
}

func clean() {
	os.Remove("test.db")
}

func TestStore(t *testing.T) {
	db := setup(t)
	defer clean()

	torrentDAO := TorrentDAO{db}
	expected := app.Torrent{
		Link:     "test-link",
		InfoHash: "test-hash",
	}
	returnValue, err := torrentDAO.Store(expected)

	if err != nil {
		t.Fatalf("failed to store: %s", err)
	}
	if returnValue.InfoHash != expected.InfoHash {
		t.Fatalf("Return value not same as expected(respective): \n%+v\n%+v", returnValue, expected)
	}

	var actual app.Torrent
	db.First(&actual, returnValue.ID)

	if expected.InfoHash != actual.InfoHash {
		t.Fatalf("Actual value in db not same as expected (respective): \n%+v\n%+v", actual, expected)
	}
}

func TestGetByID(t *testing.T) {
	db := setup(t)
	defer clean()

	torrentDAO := TorrentDAO{db}
	expected := app.Torrent{
		Link:     "test-link",
		InfoHash: "test-hash",
	}

	returnValue, err := torrentDAO.Store(expected)
	if err != nil {
		t.Fatalf("failed to store: %s", err)
	}

	actual, err := torrentDAO.GetByID(returnValue.ID)
	if err != nil {
		t.Fatalf("failed to store: %s", err)
	}
	if actual.InfoHash != expected.InfoHash {
		t.Fatalf("Actual value in db not same as expected (respective): \n%+v\n%+v", actual, expected)
	}
}

func TestGet(t *testing.T) {
	db := setup(t)
	defer clean()

	torrentDAO := TorrentDAO{db}
	expected := []app.Torrent{
		app.Torrent{
			Link:     "test-link",
			InfoHash: "test-hash",
		},
		app.Torrent{
			Link:     "test-link1",
			InfoHash: "test-hash2",
		},
	}

	for _, torrent := range expected {
		_, err := torrentDAO.Store(torrent)
		if err != nil {
			t.Fatalf("failed to store: %s", err)
		}
	}

	actual, err := torrentDAO.Get()
	if err != nil {
		t.Fatalf("failed to store: %s", err)
	}
	if len(actual) != len(expected) {
		t.Fatalf("Actual value in db not same as expected (respective): \n%+v\n%+v", actual, expected)
	}
	for i, torrent := range actual {
		if torrent.InfoHash != actual[i].InfoHash {
			t.Fatalf("Actual value in db not same as expected (respective): \n%+v\n%+v", actual, expected)
		}
	}
}
