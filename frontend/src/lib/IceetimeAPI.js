export class TorrentsAPI {
  static baseURL = "http://localhost:8080/v1";
  // ~=~=~=~=~=~=~=~=~=~=~=
  // API Endpoints
  // ~=~=~=~=~=~=~=~=~=~=~=

  static async NewMagnet(magnet) {
    return asyncApiCall(`${this.baseURL}/torrents/new/magnet`, {
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
    return asyncApiCall(`${this.baseURL}/torrents`);
  }

  static async GetTorrentByInfoHash(infoHash) {
    return asyncApiCall(`${this.baseURL}/torrents/torrent/${infoHash}`);
  }

  static async GetTorrentByInfoHash(infoHash) {
    return asyncApiCall(`${this.baseURL}/torrents/torrent/${infoHash}`);
  }

  static async FindTorrentForMovie(imdbID, title, year, minQualityIndex) {
    return asyncApiCall(`${this.baseURL}/torrents/find_for_movie?imdb_id=${imdbID}&title=${title}&year=${year}&min_quality=${minQualityIndex}`);
  }

  // ~=~=~=~=~=~=~=~=~=~=~=
  // URL Composition
  // ~=~=~=~=~=~=~=~=~=~=~=

  static ComposeURLForDirectTorrentStream(infoHash, file) {
    return `${this.baseURL}/torrents/torrent/${infoHash}/stream/${file}`
  }

}

export class TranscoderAPI {
  static baseURL = "http://localhost:8080/v1";

  static GetMetadataForFile(infoHash, file) {
    const fileURL = TorrentsAPI.ComposeURLForDirectTorrentStream(infoHash, file)
    return asyncApiCall(`${this.baseURL}/transcoder/from_url/metadata?url=${fileURL}`);
  }

  static ComposeURLForTranscodedTorrentStream(infoHash, file, resolution, maxBitrate) {
    const fileURL = TorrentsAPI.ComposeURLForDirectTorrentStream(infoHash, file)
    return `${this.baseURL}/transcoder/from_url?url=${fileURL}&res=${resolution}&max_bitrate=${maxBitrate}`;
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
