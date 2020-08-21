import videojs from 'video.js';
const Plugin = videojs.getPlugin('plugin');

export default class TimeRangesSeeking extends Plugin {
  constructor(player, options) {
    super(player, options);
    console.log('player plugin timerangeseeking');

    player.ready(function(){
      console.log('player plugin timerangeseeking player', player);
      console.log(this.currentTime.toString())

      var baseSrc = player.src();

      var seekBar = player.controlBar.progressControl.seekBar;
      seekBar._timeOffset = 0;
      seekBar._seekTime = null;

      seekBar.handleMouseMove = function(event) {
        let newTime = this.calculateDistance(event) * this.player_.duration();
        console.log("Newtime: ", newTime)

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
        console.log("offest: ", this._timeOffset, this._seekTime)

        // this._timeOffset = this._seekTime;
        // this._timeOffset = 0;
        player.play();

        this._seekTime = null;
        // player.currentTime(this._timeOffset)
      }
      _seekBarHandleMouseUp.bind(seekBar);
      seekBar.handleMouseUp.bind(seekBar);




      // Get current time, add time offset
    //   console.log(seekBar.getCurrentTime_.toString())
    //   var _seekBarGetCurrentTime_ = seekBar.getCurrentTime_;
    //   seekBar.getCurrentTime_ = function() {
    //     return ((this.player_.scrubbing()) ?
    //       this.player_.getCache().currentTime :
    //       this.player_.currentTime()) + (this._timeOffset || 0);
    //   }
    //   _seekBarGetCurrentTime_.bind(seekBar);
    //   seekBar.getCurrentTime_.bind(seekBar);
    });
  }
}
