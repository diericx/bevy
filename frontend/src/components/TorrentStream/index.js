import React from 'react';
import { Redirect } from 'react-router-dom';
import Button from 'react-bootstrap/Button';
import VideoPlayer from '../VideoPlayer';
import Row from 'react-bootstrap/Row';
import Alert from 'react-bootstrap/Alert';
import Spinner from 'react-bootstrap/Spinner';
import Col from 'react-bootstrap/Col';
import { TorrentsAPI, TranscoderAPI } from '../../lib/IceetimeAPI';
import Torrents from '../../pages/torrents';

export default class MyComponent extends React.Component {
  state = {
    torrentLink: null,
    isLoading: false,
    error: null,
  };

  async findTorrent(imdbID, title, year) {
    const resp = await TorrentsAPI.FindTorrentForMovie(imdbID, title, year, 0);
    this.setState({
      isLoading: false,
      ...resp,
    });
  }

  render() {
    const { movie } = this.props;
    const { torrentLink, isLoading, error } = this.state;

    if (error) {
      return (
        <Alert variant={'danger'} style={{ width: '80%' }}>
          Error: {error.message}
        </Alert>
      );
    }

    if (isLoading) {
      return (
        <Row
          style={{ textAlign: 'center' }}
          className={'justify-content-center align-items-center'}
        >
          <Col sm={12}>
            <p>Searching indexers for movie...</p>
          </Col>
          <Col sm={12}>
            <Spinner animation="border" role="status">
              <span className="sr-only">Loading...</span>
            </Spinner>
          </Col>
        </Row>
      );
    }

    if (!torrentLink) {
      return (
        <Button
          variant="primary"
          onClick={async () => {
            this.findTorrent(movie.imdb_id, movie.title, movie.release_year);
            this.setState({ isLoading: true });
          }}
        >
          Watch Movie
        </Button>
      );
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
