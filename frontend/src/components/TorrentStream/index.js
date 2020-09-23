import React from "react";
import { Redirect } from "react-router-dom";
import Button from "react-bootstrap/Button";
import VideoPlayer from "../VideoPlayer";
import Row from "react-bootstrap/Row";
import Alert from "react-bootstrap/Alert";
import Spinner from "react-bootstrap/Spinner";
import Col from "react-bootstrap/Col";
import { TorrentsAPI, TranscoderAPI } from "../../lib/IceetimeAPI";
import Torrents from "../../pages/torrents";

export default class MyComponent extends React.Component {
  state = {
    torrentLink: null,
    isLoading: false,
    error: null,
  };

  async findTorrent(imdbID, title, year) {
    // TODO: try catch here to handle network errors
    const resp = await TorrentsAPI.FindTorrentForMovie(imdbID, title, year, 0);
    console.log(resp);
    this.setState({
      isLoading: false,
      ...resp,
    });
  }

  render() {
    const { movie } = this.props;
    const { torrentLink, isLoading, error } = this.state;

    let releaseDate = movie.release_date.split("-")[0];

    if (error) {
      return (
        <Alert variant={"danger"} style={{ width: "80%" }}>
          Error: {error.message}
        </Alert>
      );
    }

    if (isLoading) {
      return (
        <Row
          style={{ textAlign: "center" }}
          className={"justify-content-center align-items-center"}
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
            this.findTorrent(
              movie.externalIDs.imdb_id,
              movie.title,
              releaseDate
            );
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
      width: 720,
      plugins: {
        timeRangesSeeking: {},
        durationFromServer: {},
      },
      sources: [
        {
          src: TranscoderAPI.ComposeURLForTranscodedTorrentStream(
            torrentLink.torrentInfoHash,
            torrentLink.fileIndex,
            "iw:ih",
            "1G"
          ),
          type: "video/mp4",
          label: "Original",
          selected: true,
        },
        {
          src: TranscoderAPI.ComposeURLForTranscodedTorrentStream(
            torrentLink.torrentInfoHash,
            torrentLink.fileIndex,
            "-2:1080",
            "2M"
          ),
          type: "video/mp4",
          label: "1080p",
        },
        {
          src: TranscoderAPI.ComposeURLForTranscodedTorrentStream(
            torrentLink.torrentInfoHash,
            torrentLink.fileIndex,
            "-2:720",
            "1M"
          ),
          type: "video/mp4",
          label: "720p",
        },
        {
          src: TranscoderAPI.ComposeURLForTranscodedTorrentStream(
            torrentLink.torrentInfoHash,
            torrentLink.fileIndex,
            "-2:480",
            "1M"
          ),
          type: "video/mp4",
          label: "480p",
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
