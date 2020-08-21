import React from 'react';
import videojs from 'video.js';
import TimeRangesSeeking from '../VideojsPlugins/TimeRangesSeeking';
import DurationFromServer from '../VideojsPlugins/DurationFromServer';
require('../VideojsPlugins/videojs-quality-selector/src/js')(videojs);

export default class VideoPlayer extends React.Component {
  componentDidMount() {
    // TODO: Maybe this can be done not on mount, somwhere onece?
    videojs.registerPlugin("timeRangesSeeking", TimeRangesSeeking);
    videojs.registerPlugin("durationFromServer", DurationFromServer);

    // instantiate Video.js
    this.player = videojs(this.videoNode, this.props, function onPlayerReady() {
      console.log('onPlayerReady', this)
      // this.on('qualityRequested', function(event, newSource) {
      //   this.selectedSrc = {}
      //   console.log("Quality requested", event, newSource)
      //   var seekBar = this.controlBar.progressControl.seekBar;
      //   var time = seekBar._timeOffset + this.currentTime();
      //   console.log(seekBar._timeOffset, this.currentTime(), time);
      //   let newSources = this.currentSources().map((src) => {
      //     let url = new URL(src.src)
      //     var search_params = url.searchParams;
      //     search_params.set("time", time);
      //     url.search = search_params.toString();
      //     src.src = url.toString();
      //     return src;
      //   })
      //   console.log("New sources: ", newSources)
      //   // this.src(newSources)
      // })
      // this.on('qualitySelected', function() {
      //   console.log('quality selected event start')
      //   // this.src(newSources);

      //   // this.src(
      //       // this.currentSources().map((src) => {
      //       //     let url = new URL(src.src)
      //       //     var search_params = url.searchParams;
      //       //     search_params.set("time", this._seekTime);
      //       //     url.search = search_params.toString();
      //       //     src.src = url.toString();
      //       //     return src;
      //       // })
      //   // );
      //   // this._timeOffset = this._seekTime;

      //   // this.play();

      //   // this._seekTime = null;
      // })
      this.on('timeupdate', function () {
        console.log(JSON.stringify(this.currentSources()));
        console.log(this.currentTime())
      })
      this.controlBar.addChild('QualitySelector');
    });


  }

  // destroy player on unmount
  componentWillUnmount() {
    if (this.player) {
      this.player.dispose()
    }
  }

  // wrap the player in a div with a `data-vjs-player` attribute
  // so videojs won't create additional wrapper in the DOM
  // see https://github.com/videojs/video.js/pull/3856
  render() {
    return (
      <div>	
        <div data-vjs-player>
          <video ref={ node => this.videoNode = node } className="video-js"></video>
        </div>
      </div>
    )
  }
}
