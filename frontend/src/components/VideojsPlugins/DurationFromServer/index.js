import videojs from 'video.js';
const Plugin = videojs.getPlugin('plugin');

export default class DurationFromServer extends Plugin { 
  constructor(player, options) {
    super(player, options);

    var plugin = this;

    this._cachedDurations = {};
    
    player.ready(function(){  
      plugin.setPlayerDurationFromServer(player);
      
      player.on('durationchange', function(event) {
        plugin.setPlayerDurationFromServer(player);
      });
    });
  }

  getBaseURL(url) {
      return url.split('?')[0];
  }

  getCachedDuration(baseURL) {
    if(baseURL in this._cachedDurations) {
      return this._cachedDurations[baseURL];
    }
    return null;
  }

  setCachedDuration(baseURL, duration) {
    this._cachedDurations[baseURL] = duration;
    return this._cachedDurations[baseURL];
  }

  setPlayerDurationFromServer(player) {
    this.getDuration(player.src()).then(function(duration) {
      if(player.duration() !== duration) {
        player.duration(duration);
      }
    });
  }
  
  getDuration(url) {
    var baseURL = this.getBaseURL(url);
    var duration = this.getCachedDuration(baseURL);
    if(duration !== null) {
      return Promise.resolve(duration);
    }

    var plugin = this;

    console.log("Get duration...")

    return fetch(baseURL+'/metadata')
    .then((resp) => resp.json())
    .then(function(metadata) {
      return plugin.setCachedDuration(url, metadata.format.duration);
    });
  }
}