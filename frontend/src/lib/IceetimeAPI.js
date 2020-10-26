export class TorrentsAPI {
  static backendURL = window._env_.BACKEND_URL;
  // ~=~=~=~=~=~=~=~=~=~=~=
  // API Endpoints
  // ~=~=~=~=~=~=~=~=~=~=~=

  static async NewMagnet(magnet) {
    return asyncApiCall(`${this.backendURL}/v1/torrents/new/magnet`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        magnet_url: magnet,
      }),
    });
  }

  static async Get() {
    return asyncApiCall(`${this.backendURL}/v1/torrents`);
  }

  static async GetTorrentByInfoHash(infoHash) {
    return asyncApiCall(`${this.backendURL}/v1/torrents/torrent/${infoHash}`);
  }

  static async GetTorrentByInfoHash(infoHash) {
    return asyncApiCall(`${this.backendURL}/v1/torrents/torrent/${infoHash}`);
  }

  static async FindTorrentForMovie(imdbID, title, year, minQualityIndex) {
    return asyncApiCall(
      `${this.backendURL}/v1/torrents/find_for_movie?imdb_id=${imdbID}&title=${title}&year=${year}&min_quality=${minQualityIndex}`
    );
  }

  // ~=~=~=~=~=~=~=~=~=~=~=
  // URL Composition
  // ~=~=~=~=~=~=~=~=~=~=~=

  static ComposeURLForDirectTorrentStream(infoHash, file) {
    return `${this.backendURL}/v1/torrents/torrent/${infoHash}/stream/${file}`;
  }
}

export class TranscoderAPI {
  static backendURL = window._env_.BACKEND_URL;

  static GetMetadataForFile(infoHash, file) {
    const fileURL = TorrentsAPI.ComposeURLForDirectTorrentStream(
      infoHash,
      file
    );
    return asyncApiCall(
      `${this.backendURL}/v1/transcoder/from_url/metadata?url=${fileURL}`
    );
  }

  static ComposeURLForTranscodedTorrentStream(
    infoHash,
    file,
    resolution,
    maxBitrate
  ) {
    const fileURL = TorrentsAPI.ComposeURLForDirectTorrentStream(
      infoHash,
      file
    );
    return `${this.backendURL}/v1/transcoder/from_url?url=${fileURL}&res=${resolution}&max_bitrate=${maxBitrate}`;
  }
}

export class TmdbAPI {
  static backendURL = window._env_.BACKEND_URL;

  static PopularMovies() {
    return asyncApiCall(`${this.backendURL}/v1/tmdb/browse/movies/popular`);
  }

  static SearchMovie(query) {
    return asyncApiCall(
      `${this.backendURL}/v1/tmdb/search/movies?query=${query}`
    );
  }

  static GetMovie(id) {
    return asyncApiCall(`${this.backendURL}/v1/tmdb/movies/${id}`);
  }
}
async function asyncApiCall(url, options) {
  try {
    const resp = await fetch(url, options);
    const json = await resp.json();
    return json;
  } catch (error) {
    return {
      error,
    };
  }
}
