package services

import (
	"testing"
	"time"

	"github.com/diericx/iceetime/internal/app/mocks"
	"github.com/golang/mock/gomock"
)

type MockState struct {
	torrentClient *mocks.MockTorrentClient
	torrent       Torrent
}

func setup(ctrl *gomock.Controller) MockState {
	mockTorrentClient := mocks.NewMockTorrentClient(ctrl)
	return MockState{
		torrentClient: mockTorrentClient,
		torrent: Torrent{
			Client:         mockTorrentClient,
			GetInfoTimeout: time.Second * 1,
		},
	}
}

func TestAddFromMagnet(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := setup(ctrl)

	mockTorrent := mocks.NewMockTorrent(ctrl)
	magnet := "test-magnet"

	gotInfoChan := make(chan struct{})
	go func() {
		gotInfoChan <- struct{}{}
	}()

	mockTorrent.EXPECT().GotInfo().Return(gotInfoChan)
	s.torrentClient.EXPECT().AddMagnet(magnet).Return(mockTorrent, nil)

	_, err := s.torrent.AddFromMagnet(magnet)
	if err != nil {
		t.Fatalf("Unexepcted error: %v", err)
	}

}
