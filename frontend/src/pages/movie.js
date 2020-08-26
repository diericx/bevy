import React from "react";
import { Redirect } from "react-router-dom";
import Button from "react-bootstrap/Button";
import VideoPlayer from "../components/VideoPlayer";
import Row from "react-bootstrap/Row";
import Container from "react-bootstrap/Container";
import Spinner from "react-bootstrap/Spinner";
import Col from "react-bootstrap/Col";
import "./movie.css";
export default class MyComponent extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      movie: null,
      torrent: null,
      isTorrentLoading: false,
    };
  }

  componentDidMount() {
    let {
      location: {
        state: { movie },
      },
    } = this.props;
    if (!movie.externalIDs) {
      let tmdbAPIKey = process.env.REACT_APP_TMDB_API_KEY;
      fetch(
        `https://api.themoviedb.org/3/movie/${movie.id}/external_ids?api_key=${tmdbAPIKey}`
      )
        .then((res) => res.json())
        .then(
          (result) => {
            this.setState({
              isLoaded: true,
              movie: {
                ...movie,
                externalIDs: result,
              },
            });
          },
          (error) => {
            this.setState({
              isLoaded: true,
              error,
            });
          }
        );
    } else {
      this.setState({
        isLoaded: true,
        movie: movie,
      });
    }
  }

  findTorrent(imdbID, title, year) {
    fetch(
      `http://localhost:8080/find/movie?imdbid=${imdbID}&title=${title}&year=${year}`
    )
      .then((res) => res.json())
      .then(
        (result) => {
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

  renderTorrentStreamingOptions() {
    const { movie, torrent, isTorrentLoading } = this.state;
    let releaseDate = movie.release_date.split("-")[0];

    if (isTorrentLoading) {
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
          src: `http://localhost:8080/stream/torrent/${torrent.id}/transcode`,
          type: "video/mp4",
          label: "Original",
          selected: true,
        },
        {
          src: `http://localhost:8080/stream/torrent/${torrent.id}/transcode?res=-2:1080&max_bitrate=2M`,
          type: "video/mp4",
          label: "1080p",
        },
        {
          src: `http://localhost:8080/stream/torrent/${torrent.id}/transcode?res=-2:720&max_bitrate=1M`,
          type: "video/mp4",
          label: "720p",
        },
        {
          src: `http://localhost:8080/stream/torrent/${torrent.id}/transcode?res=-2:480&max_bitrate=1M`,
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

  render() {
    const { error, isLoaded, movie, torrent, isTorrentLoading } = this.state;

    if (error) {
      return <div>Error: {error.message}</div>;
    } else if (!isLoaded) {
      return <div>Loading...</div>;
    }

    return (
      <Container fluid>
        <Row
          style={{
            backgroundImage: `url("https://image.tmdb.org/t/p/original${movie.backdrop_path}")`,
          }}
          className={"movie-banner-row"}
        >
          <div className={"movie-banner-row-filler"}>
            <Col>
              <Row className={"align-items-center"}>
                <Col sm={12} md={3}>
                  <img
                    className={"movie-poster"}
                    src={`https://image.tmdb.org/t/p/w500${movie.poster_path}`}
                  ></img>
                </Col>
                <Col sm={12} md={6} className={"movie-details"}>
                  <h1>{movie.title}</h1>
                  <p>{movie.vote_average}</p>
                  <h2>Overview</h2>
                  <p>{movie.overview}</p>
                </Col>
              </Row>
            </Col>
          </div>
        </Row>
        <br />
        <br />
        <Row className={"justify-content-center"}>
          {this.renderTorrentStreamingOptions()}
        </Row>
        <br />
        <br />
      </Container>
    );
  }
}
