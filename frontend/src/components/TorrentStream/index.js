
import React from "react";
import { Redirect } from "react-router-dom";
import Button from "react-bootstrap/Button";
import VideoPlayer from "../VideoPlayer";
import Row from "react-bootstrap/Row";
import Alert from "react-bootstrap/Alert";
import Spinner from "react-bootstrap/Spinner";
import Col from "react-bootstrap/Col";

let backendURL = window._env_.BACKEND_URL;

export default class MyComponent extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      torrent: null,
      isLoading: false,
      error: null,
    };
  }

  findTorrent(imdbID, title, year) {
    fetch(
      `http://${backendURL}/find/movie?imdbid=${imdbID}&title=${title}&year=${year}`
    )
      .then((res) => res.json())
      .then(
        (result) => {
          if (result && result.error) {
            this.setState({
              isTorrentLoading: false,
              error: result,
            });
            return
          }

          this.setState({
            isTorrentLoading: false,
            torrent: result,
          });
        },
        (error) => {
          this.setState({
            isTorrentLoading: false,
            error,
          });
        }
      );
  }

  render() {
    const { movie } = this.props;
    const { torrent, isLoading, error } = this.state;

    let releaseDate = movie.release_date.split("-")[0];

    if (error) {
      return <Alert variant={'danger'} style={{width: "80%"}}>
        Error: {error.message}
      </Alert>
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

    if (!torrent) {
      return (
        <Button
          variant="primary"
          onClick={() => {
            this.findTorrent(
              movie.externalIDs.imdb_id,
              movie.title,
              releaseDate
            );
            this.setState({ isTorrentLoading: true });
          }}
        >
          Watch Movie
        </Button>
      );
    }


    const videoJsOptions = {
      autoplay: true,
      controls: true,
      width: 720,
      plugins: {
        timeRangesSeeking: {},
        durationFromServer: {},
      },
      sources: [
        {
          src: `http://${backendURL}/stream/torrent/${torrent.id}/transcode`,
          type: "video/mp4",
          label: "Original",
          selected: true,
        },
        {
          src: `http://${backendURL}/stream/torrent/${torrent.id}/transcode?res=-2:1080&max_bitrate=2M`,
          type: "video/mp4",
          label: "1080p",
        },
        {
          src: `http://${backendURL}/stream/torrent/${torrent.id}/transcode?res=-2:720&max_bitrate=1M`,
          type: "video/mp4",
          label: "720p",
        },
        {
          src: `http://${backendURL}/stream/torrent/${torrent.id}/transcode?res=-2:480&max_bitrate=1M`,
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
