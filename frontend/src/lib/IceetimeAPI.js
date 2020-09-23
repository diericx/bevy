export class TorrentsAPI {
  static backendURL = window._env_.BACKEND_URL;
  // ~=~=~=~=~=~=~=~=~=~=~=
  // API Endpoints
  // ~=~=~=~=~=~=~=~=~=~=~=

  static async NewMagnet(magnet) {
    return asyncApiCall(`${this.backendURL}/torrents/new/magnet`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        magnet_url: magnet,
      }),
    });
  }

  static async Get() {
    return asyncApiCall(`${this.backendURL}/torrents`);
  }

  static async GetTorrentByInfoHash(infoHash) {
    return asyncApiCall(`${this.backendURL}/torrents/torrent/${infoHash}`);
  }

  static async GetTorrentByInfoHash(infoHash) {
    return asyncApiCall(`${this.backendURL}/torrents/torrent/${infoHash}`);
  }

  static async FindTorrentForMovie(imdbID, title, year, minQualityIndex) {
    return asyncApiCall(
      `${this.backendURL}/torrents/find_for_movie?imdb_id=${imdbID}&title=${title}&year=${year}&min_quality=${minQualityIndex}`
    );
  }

  // ~=~=~=~=~=~=~=~=~=~=~=
  // URL Composition
  // ~=~=~=~=~=~=~=~=~=~=~=

  static ComposeURLForDirectTorrentStream(infoHash, file) {
    return `${this.backendURL}/torrents/torrent/${infoHash}/stream/${file}`;
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
      `${this.backendURL}/transcoder/from_url/metadata?url=${fileURL}`
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
    return `${this.backendURL}/transcoder/from_url?url=${fileURL}&res=${resolution}&max_bitrate=${maxBitrate}`;
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
