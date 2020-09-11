package torrent

import (
	"github.com/anacrolix/torrent"
	"github.com/diericx/iceetime/internal/app"
)

func AnacrolixTorrentToApp(t *torrent.Torrent) app.Torrent {
	return app.Torrent{
		InfoHash: t.InfoHash(),
		Stats:    t.Stats(),
		Length:   t.Length(),
		Name:     t.Name(),
		Seeding:  t.Seeding(),
	}

}
