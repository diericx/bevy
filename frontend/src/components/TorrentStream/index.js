import React from 'react';
import VideoPlayer from '../VideoPlayer';
import Alert from 'react-bootstrap/Alert';
import { TranscoderAPI } from '../../lib/BevyAPI';

export default class MyComponent extends React.Component {
  state = {
    torrentLink: null,
    isLoading: false,
    error: null,
  };

  render() {
    const { movie, torrentLink } = this.props;
    const { isLoading, error } = this.state;

    if (error) {
      return (
        <Alert variant={'danger'} style={{ width: '80%' }}>
          Error: {error.message}
        </Alert>
      );
    }

    if (!torrentLink) {
      return null;
    }

    const videoJsOptions = {
      infoHash: torrentLink.torrentInfoHash,
      fileIndex: torrentLink.fileIndex,
      autoplay: true,
      controls: true,
      techOrder: ['chromecast', 'html5'], // You may have more Tech, such as Flash or HLS
      width: 720,
      plugins: {
        timeRangesSeeking: {},
        durationFromServer: {},
        chromecast: {
          addButtonToControlBar: true,
        },
      },
      sources: [
        {
          src: TranscoderAPI.ComposeURLForTranscodedTorrentStream(
            torrentLink.torrentInfoHash,
            torrentLink.fileIndex,
            'iw:ih',
            '1G'
          ),
          type: 'video/mp4',
          label: 'Original',
          selected: true,
        },
        {
          src: TranscoderAPI.ComposeURLForTranscodedTorrentStream(
            torrentLink.torrentInfoHash,
            torrentLink.fileIndex,
            '-2:1080',
            '2M'
          ),
          type: 'video/mp4',
          label: '1080p',
        },
        {
          src: TranscoderAPI.ComposeURLForTranscodedTorrentStream(
            torrentLink.torrentInfoHash,
            torrentLink.fileIndex,
            '-2:720',
            '1M'
          ),
          type: 'video/mp4',
          label: '720p',
        },
        {
          src: TranscoderAPI.ComposeURLForTranscodedTorrentStream(
            torrentLink.torrentInfoHash,
            torrentLink.fileIndex,
            '-2:480',
            '1M'
          ),
          type: 'video/mp4',
          label: '480p',
        },
      ],
    };

    return (
      <div>
        <VideoPlayer {...videoJsOptions} />
      </div>
    );
  }
}
