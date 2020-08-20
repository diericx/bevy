import videojs from 'video.js';
const Plugin = videojs.getPlugin('plugin');

export default class TimeRangesSeeking extends Plugin {
  constructor(player, options) {
    super(player, options);
    console.log('player plugin timerangeseeking');
    
    player.ready(function(){
      console.log('player plugin timerangeseeking player', player);
      
      var baseSrc = player.src();

      var seekBar = player.controlBar.progressControl.seekBar;
      seekBar._timeOffset = 0;
      seekBar._seekTime = null;

      
      /*var _seekBarHandleMouseDown = seekBar.handleMouseDown;
      seekBar.handleMouseDown = function(event) {
        _seekBarHandleMouseDown(event);
        this._seekTime = null;
      }
      seekBar.handleMouseDown.bind(seekBar);*/
      
      // Handle mouse move, to request new time based content from server
      var _seekBarHandleMouseMove = seekBar.handleMouseMove;
      seekBar.handleMouseMove = function(event) {
        let newTime = this.calculateDistance(event) * this.player_.duration();

        // Don't let video end while scrubbing.
        if (newTime === this.player_.duration()) {
          newTime = newTime - 0.1;
        }

        this._seekTime = newTime;
        this._timeOffset = this._seekTime;
      }
      _seekBarHandleMouseMove.bind(seekBar);
      seekBar.handleMouseMove.bind(seekBar);

      var _seekBarHandleMouseUp = seekBar.handleMouseUp;
      seekBar.handleMouseUp = function(event) {
        _seekBarHandleMouseUp.bind(seekBar)(event);

        if(this._seekTime === null) {
          return;
        }

        console.log('set new src', baseSrc+'?time='+this._seekTime);
        this._timeOffset = this._seekTime;
        player.src({type: 'video/mp4', src: baseSrc+'?time='+this._seekTime});
        player.play();

        this._seekTime = null;
      }
      _seekBarHandleMouseUp.bind(seekBar);
      seekBar.handleMouseUp.bind(seekBar);
      
      // Get current time, add time offset
      var _seekBarGetCurrentTime_ = seekBar.getCurrentTime_;
      seekBar.getCurrentTime_ = function() {
        console.log("Get currnet time {")
        console.log(this.player_.scrubbing())
        console.log(this.player_.getCache().currentTime)
        console.log(this.player_.currentTime() );
        console.log((this._timeOffset || 0))
        console.log(((this.player_.scrubbing()) ?
          this.player_.getCache().currentTime :
          this.player_.currentTime()) + (this._timeOffset || 0))
        return ((this.player_.scrubbing()) ?
          this.player_.getCache().currentTime :
          this.player_.currentTime()) + (this._timeOffset || 0);
      }
      _seekBarGetCurrentTime_.bind(seekBar);
      seekBar.getCurrentTime_.bind(seekBar);
    });

    /*player.on("timeupdate", function(event) {
      //console.log('timeupdate', player.currentTime());
    });
    
    player.on("seeking", function(event) {
      console.log('seeking', player.currentTime());
    });

    player.on("seeked", function(event) {
      console.log('seeked', player.currentTime());
    });

    player.on('playing', function() {
      videojs.log('playback began!');
    });*/
  }
}
