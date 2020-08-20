import React from 'react';
import videojs from 'video.js';
import TimeRangesSeeking from '../VideojsPlugins/TimeRangesSeeking';
import DurationFromServer from '../VideojsPlugins/DurationFromServer';

export default class VideoPlayer extends React.Component {
  componentDidMount() {
    // TODO: Maybe this can be done not on mount, somwhere onece?
    videojs.registerPlugin("timeRangesSeeking", TimeRangesSeeking);
    videojs.registerPlugin("durationFromServer", DurationFromServer);

    // instantiate Video.js
    this.player = videojs(this.videoNode, this.props, function onPlayerReady() {
      console.log('onPlayerReady', this)
    });
    // player.timeRangesSeeking();
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
