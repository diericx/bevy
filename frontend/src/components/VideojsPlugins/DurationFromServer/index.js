import videojs from 'video.js';
const Plugin = videojs.getPlugin('plugin');

export default class DurationFromServer extends Plugin { 
  static _cachedDurations = {};

  constructor(player, options) {
    super(player, options);

    var plugin = this;

    player.ready(function(){  
      plugin.setPlayerDurationFromServer(player);
      
      player.on('durationchange', function(event) {
        plugin.setPlayerDurationFromServer(player);
      });
    });
  }

  getCachedDuration(url) {
    if(url.pathname in DurationFromServer._cachedDurations) {
      return DurationFromServer._cachedDurations[url.pathname];
    }
    return null;
  }

  setCachedDuration(url, duration) {
    DurationFromServer._cachedDurations[url.pathname] = duration;
    return DurationFromServer._cachedDurations[url.pathname];
  }

  setPlayerDurationFromServer(player) {
    this.getDuration(player.src()).then(function(duration) {
      if(player.duration() !== duration) {
        player.duration(duration);
      }
    });
  }
  
  getDuration(urlString) {
    let url = new URL(urlString);
    var duration = this.getCachedDuration(url);
    if(duration !== null) {
      return Promise.resolve(duration);
    }

    var plugin = this;

    return fetch(url.origin+url.pathname+'/metadata')
    .then((resp) => resp.json())
    .then(function(metadata) {
      return plugin.setCachedDuration(url, metadata.format.duration);
    });
  }
}