package app

import (
	"errors"
	"fmt"
)

type MainConfig struct {
	Transcoder     TranscoderConfig     `toml:"transcoder"`
	Tmdb           TmdbConfig           `toml:"tmdb"`
	ReleaseService ReleaseServiceConfig `toml:"torrent_fetcher"`
	TorrentClient  TorrentClientConfig  `toml:"torrent_client"`
}

type TorrentClientConfig struct {
	TorrentInfoTimeout                int    `toml:"info_timeout"`
	TorrentFilePath                   string `toml:"file_path"`
	TorrentDataPath                   string `toml:"data_path"`
	TorrentHalfOpenConnsPerTorrent    int    `toml:"half_open_conns_per_torrent"`
	TorrentEstablishedConnsPerTorrent int    `toml:"established_conns_per_torrent"`
	MetaRefreshRate                   int    `toml:"meta_refresh_rate"`
}

type TmdbConfig struct {
	APIKey string `toml:"api_key"`
}

type TranscoderConfig struct {
	Video struct {
		Format          string `toml:"format"`
		CompressionAlgo string `toml:"compression_algo"`
	} `toml:"video"`
	Audio struct {
		CompressionAlgo string `toml:"compression_algo"`
	} `toml:"audio"`
}

type ReleaseServiceConfig struct {
	Indexers         []Indexer `toml:"indexers"`
	Qualities        []Quality `toml:"qualities"`
	QualityScoreExpr string    `toml:"quality_score_expr"`
}

func (c ReleaseServiceConfig) Validate() error {
	if len(c.Indexers) == 0 {
		return errors.New("Indexers array must have length of at least 1")
	}
	for i, indexer := range c.Indexers {
		if err := indexer.Validate(); err != nil {
			return fmt.Errorf("Indexer %v is invalid:\n%+v\n %s", i, indexer, err)
		}
	}
	if len(c.Qualities) == 0 {
		return errors.New("Indexers array must have length of at least 1")
	}
	for i, quality := range c.Qualities {
		if err := quality.Validate(); err != nil {
			return fmt.Errorf("Quality %v is invalid:\n%+v\n %s", i, quality, err)
		}
	}
	if c.QualityScoreExpr == "" {
		return errors.New("Size score function cannot be empty")
	}
	return nil
}

func (c MainConfig) Validate() error {
	if err := c.Transcoder.Validate(); err != nil {
		return fmt.Errorf("Transcoder is invalid: %s", err)
	}
	if err := c.Tmdb.Validate(); err != nil {
		return fmt.Errorf("Tmdb is invalid: %s", err)
	}
	if err := c.TorrentClient.Validate(); err != nil {
		return fmt.Errorf("TorrentClient is invalid: %s", err)
	}
	if err := c.ReleaseService.Validate(); err != nil {
		return fmt.Errorf("Release Service is invalid: %s", err)
	}
	return nil
}

func (c TranscoderConfig) Validate() error {
	if c.Video.Format == "" {
		return errors.New("Video format cannot be empty string")
	}
	if c.Video.CompressionAlgo == "" {
		return errors.New("Video compression algorithm cannot be empty string")
	}
	if c.Audio.CompressionAlgo == "" {
		return errors.New("Audio compression algorithm cannot be empty string")
	}
	return nil
}

func (c TmdbConfig) Validate() error {
	if c.APIKey == "" {
		return errors.New("API Key cannot be emtpy string")
	}
	return nil
}

func (c TorrentClientConfig) Validate() error {
	if c.TorrentInfoTimeout == 0 {
		return errors.New("Torrent info timeout cannot be 0")
	}
	if c.TorrentFilePath == "" {
		return errors.New("TorrentFilePath cannot be empty")
	}
	if c.TorrentDataPath == "" {
		return errors.New("TorrentDataPath cannot be empty")
	}
	if c.TorrentHalfOpenConnsPerTorrent+c.TorrentEstablishedConnsPerTorrent == 0 {
		return errors.New("(Half open connections per torrent + established connections per torrent) should not equal 0")
	}
	if c.MetaRefreshRate < 1 {
		return errors.New("Meta refresh rate cannot be less than 1")
	}
	return nil
}
