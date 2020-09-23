import React from 'react';
import videojs from 'video.js';
import TimeRangesSeeking from '../VideojsPlugins/TimeRangesSeeking';
import DurationFromServer from '../VideojsPlugins/DurationFromServer';
require('../VideojsPlugins/videojs-quality-selector/src/js')(videojs);
require('@silvermine/videojs-chromecast')(videojs, { preloadWebComponents: true });

export default class VideoPlayer extends React.Component {
  componentDidMount() {
    // TODO: Maybe this can be done not on mount, somwhere onece?
    videojs.registerPlugin("timeRangesSeeking", TimeRangesSeeking);
    videojs.registerPlugin("durationFromServer", DurationFromServer);

    let options = {
      ...this.props,
      controls: true,
      techOrder: [ 'chromecast', 'html5' ], // You may have more Tech, such as Flash or HLS
      plugins: {
         chromecast: {
          addButtonToControlBar: true
         }
      }
   };

    // instantiate Video.js
    this.player = videojs(this.videoNode, options, function onPlayerReady() {
      this.currentTime = function(seconds) {
        var seekBar = this.controlBar.progressControl.seekBar;
        if (typeof seconds !== 'undefined') {
            if (seconds < 0) {
                seconds = 0;
            }

            this.techCall_('setCurrentTime', seconds);
            return;
        } // cache last currentTime and return. default to 0 seconds
        //
        // Caching the currentTime is meant to prevent a massive amount of reads on the tech's
        // currentTime when scrubbing, but may not provide much performance benefit afterall.
        // Should be tested. Also something has to read the actual current time or the cache will
        // never get updated.
        this.cache_.currentTime = this.techGet_('currentTime') + seekBar._timeOffset || seekBar._timeOffset;
        return this.cache_.currentTime;
      }

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
    const { infoHash, fileIndex } = this.props;
    return (
      <div>
        <div data-vjs-player>
          <video ref={ node => this.videoNode = node } className="video-js" infoHash={infoHash} fileIndex={fileIndex}></video>
        </div>
      </div>
    )
  }
}
