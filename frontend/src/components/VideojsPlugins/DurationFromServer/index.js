import videojs from 'video.js';
const Plugin = videojs.getPlugin('plugin');

export default class DurationFromServer extends Plugin {
  static _cachedDurations = {};

  constructor(player, options) {
    super(player, options);

    var plugin = this;

    player.ready(() => {
      this.getDuration(player.src()).then(function(duration) {
        // Set duration once with default duration function to dispatch event
        player.duration(duration);
        // Override duration to just return this value from now on
        player.duration = function(seconds) {
          if (seconds === undefined) {
            // return NaN if the duration is not known
            return this.cache_.duration !== undefined ? this.cache_.duration : NaN;
          }
        }
      })
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
      console.log("Got duration: ", duration)
      // console.log(player.duration.toString());
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
