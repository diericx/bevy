import videojs from 'video.js';
const Plugin = videojs.getPlugin('plugin');

export default class TimeRangesSeeking extends Plugin {
  constructor(player, options) {
    super(player, options);

    player.ready(function(){
      var seekBar = player.controlBar.progressControl.seekBar;
      seekBar._timeOffset = 0;
      seekBar._seekTime = null;

      seekBar.handleMouseMove = function(event) {
        let newTime = this.calculateDistance(event) * this.player_.duration();

        // Don't let video end while scrubbing.
        if (newTime === this.player_.duration()) {
          newTime = newTime - 0.1;
        }

        this._seekTime = newTime;
        this._timeOffset = this._seekTime;
      }

      var _seekBarHandleMouseUp = seekBar.handleMouseUp;
      seekBar.handleMouseUp = function(event) {
        _seekBarHandleMouseUp.bind(seekBar)(event); // What is this??

        if(this._seekTime === null) {
          return;
        }

        player.src(
            player.currentSources().map((src) => {
                let url = new URL(src.src)
                var search_params = url.searchParams;
                search_params.set("time", this._seekTime);
                url.search = search_params.toString();
                src.src = url.toString();
                return src;
            })
        );
        player.play();

        this._seekTime = null;
      }
      _seekBarHandleMouseUp.bind(seekBar);
      seekBar.handleMouseUp.bind(seekBar);
    });
  }
}
