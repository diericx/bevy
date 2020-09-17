package services

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/diericx/iceetime/internal/app"
	"github.com/diericx/iceetime/internal/app/mocks"
	"github.com/golang/mock/gomock"
)

type MockState struct {
	torrentClient   *mocks.MockTorrentClient
	torrentMetaRepo *mocks.MockTorrentMetaRepo
	torrent         Torrent
}

func setup(ctrl *gomock.Controller) MockState {
	mockTorrentClient := mocks.NewMockTorrentClient(ctrl)
	mockTorrentMetaRepo := mocks.NewMockTorrentMetaRepo(ctrl)
	return MockState{
		torrentClient:   mockTorrentClient,
		torrentMetaRepo: mockTorrentMetaRepo,

		torrent: Torrent{
			Client:           mockTorrentClient,
			TorrentMetaRepo:  mockTorrentMetaRepo,
			GetInfoTimeout:   time.Second * 1,
			TorrentFilesPath: "./",
		},
	}
}

func teardown(t *testing.T, s MockState) {
	err := filepath.Walk(s.torrent.TorrentFilesPath, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".torrent" {
			os.Remove(path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Error on teardown: %v", err)
	}
}

func TestAddFromMagnetShouldAddToClientAndSaveToFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := setup(ctrl)
	defer teardown(t, s)

	mockTorrent := mocks.NewMockTorrent(ctrl)
	mockMetaInfo := metainfo.MetaInfo{
		InfoBytes: []byte("random"),
	}
	mockTorrentMeta := app.TorrentMeta{
		HoursToStop: 100,
		InfoHash:    mockMetaInfo.HashInfoBytes().HexString(),
	}
	magnet := "test-magnet"

	gotInfoChan := make(chan struct{})
	go func() {
		gotInfoChan <- struct{}{}
	}()

	mockTorrent.EXPECT().GotInfo().Return(gotInfoChan)
	mockTorrent.EXPECT().InfoHash().Return(mockMetaInfo.HashInfoBytes()).Times(2)
	mockTorrent.EXPECT().Metainfo().Return(mockMetaInfo)
	mockTorrent.EXPECT().DownloadAll()
	s.torrentClient.EXPECT().AddMagnet(magnet).Return(mockTorrent, nil)
	s.torrentMetaRepo.EXPECT().Store(mockTorrentMeta).Return(nil)

	_, err := s.torrent.AddFromMagnet(magnet, mockTorrentMeta)
	if err != nil {
		t.Fatalf("Unexepcted error: %v", err)
	}
}

func TestSaveTorrentToFileCreatesTorrentFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := setup(ctrl)
	defer teardown(t, s)

	mockTorrent := mocks.NewMockTorrent(ctrl)
	mockMetaInfo := metainfo.MetaInfo{
		InfoBytes: []byte("random"),
	}
	mockTorrent.EXPECT().InfoHash().Return(mockMetaInfo.HashInfoBytes()).Times(2)
	mockTorrent.EXPECT().Metainfo().Return(mockMetaInfo)

	err := s.torrent.saveTorrentToFile(mockTorrent)
	if err != nil {
		t.Fatalf("Unexepected error: %v", err)
	}

	if _, err := os.Stat(fmt.Sprintf("%s%s.torrent", s.torrent.TorrentFilesPath, mockMetaInfo.HashInfoBytes().HexString())); os.IsNotExist(err) {
		t.Fatalf("File does not exist: %v", err)
	}

}

func TestAddTorrentsOnDiskShouldAddFilesWithTorrentExtension(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := setup(ctrl)
	defer teardown(t, s)

	// Save mock torrent to file
	mockMetaInfo := metainfo.MetaInfo{
		InfoBytes: []byte("random"),
	}
	gotInfoChan := make(chan struct{})
	go func() {
		gotInfoChan <- struct{}{}
	}()
	mockTorrent := mocks.NewMockTorrent(ctrl)
	mockTorrent.EXPECT().InfoHash().Return(mockMetaInfo.HashInfoBytes())
	mockTorrent.EXPECT().Metainfo().Return(mockMetaInfo)
	mockTorrent.EXPECT().GotInfo().Return(gotInfoChan)
	s.torrent.saveTorrentToFile(mockTorrent)

	s.torrentClient.EXPECT().AddFile(fmt.Sprintf("%s.torrent", mockMetaInfo.HashInfoBytes().HexString())).Return(mockTorrent, nil)

	err := s.torrent.AddTorrentsOnDisk()
	if err != nil {
		t.Fatalf("Unexepected error: %v", err)
	}
}
