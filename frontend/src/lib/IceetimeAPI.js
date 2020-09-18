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

  static async FindTorrentForMovie(imdbID) {
    return asyncApiCall(`${this.baseURL}/torrents/find_for_movie`);
  }

  // ~=~=~=~=~=~=~=~=~=~=~=
  // URL Composition
  // ~=~=~=~=~=~=~=~=~=~=~=

  ComposeURLForTorrentStream(infoHash, file, resolution, maxBitrate) {
    return `${this.baseURL}/torrents/torrent/${infoHash}/stream/${file}?res=${resolution}&max_bitrate=${maxBitrate}`;
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
