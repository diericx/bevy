const backendURL =
  process.env.REACT_APP_BACKEND_URL || window.location.origin.toString();

export class TorrentsAPI {
  // ~=~=~=~=~=~=~=~=~=~=~=
  // API Endpoints
  // ~=~=~=~=~=~=~=~=~=~=~=

  static async NewMagnet(magnet) {
    return asyncApiCall(`${backendURL}/v1/torrents/new/magnet`, {
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
    return asyncApiCall(`${backendURL}/v1/torrents`);
  }

  static async GetTorrentByInfoHash(infoHash) {
    return asyncApiCall(`${backendURL}/v1/torrents/torrent/${infoHash}`);
  }

  static async GetTorrentByInfoHash(infoHash) {
    return asyncApiCall(`${backendURL}/v1/torrents/torrent/${infoHash}`);
  }

  static async FindTorrentForMovie(imdbID, title, year, minQualityIndex) {
    return asyncApiCall(
      `${backendURL}/v1/torrents/find_for_movie?imdb_id=${imdbID}&title=${title}&year=${year}&min_quality=${minQualityIndex}`
    );
  }

  static async ScoredReleasesForMovie(imdbID, title, year, minQualityIndex) {
    return asyncApiCall(
      `${backendURL}/v1/torrents/scored_releases_for_movie?imdb_id=${imdbID}&title=${title}&year=${year}&min_quality=${minQualityIndex}`
    );
  }

  // ~=~=~=~=~=~=~=~=~=~=~=
  // URL Composition
  // ~=~=~=~=~=~=~=~=~=~=~=

  static ComposeURLForDirectTorrentStream(infoHash, file) {
    return `${backendURL}/v1/torrents/torrent/${infoHash}/stream/${file}`;
  }
}

export class TranscoderAPI {
  static GetMetadataForFile(infoHash, file) {
    const fileURL = TorrentsAPI.ComposeURLForDirectTorrentStream(
      infoHash,
      file
    );
    return asyncApiCall(
      `${backendURL}/v1/transcoder/from_url/metadata?url=${fileURL}`
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
    return `${backendURL}/v1/transcoder/from_url?url=${fileURL}&res=${resolution}&max_bitrate=${maxBitrate}`;
  }
}

export class TmdbAPI {
  static PopularMovies() {
    return asyncApiCall(`${backendURL}/v1/tmdb/browse/movies/popular`);
  }

  static SearchMovie(query) {
    return asyncApiCall(`${backendURL}/v1/tmdb/search/movies?query=${query}`);
  }

  static GetMovie(id) {
    return asyncApiCall(`${backendURL}/v1/tmdb/movies/${id}`);
  }
}

// Handles responses from our API. Expects an error field in body when there is an issue.
// returns: { ok: boolean, ...json }
async function asyncApiCall(url, options) {
  var resp;
  try {
    resp = await fetch(url, options);
    const json = await resp.json();
    return {
      ok: resp.ok,
      ...json,
    };
  } catch (error) {
    return {
      ok: false,
      error: error.message,
    };
  }
}
