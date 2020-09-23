import React from 'react';
import Row from 'react-bootstrap/Row';
import Container from 'react-bootstrap/Container';
import Col from 'react-bootstrap/Col';
import './movie.css';
import TorrentStream from '../components/TorrentStream';

let backendURL = window._env_.BACKEND_URL;

export default class MyComponent extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      movie: null,
      torrent: null,
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

  render() {
    const { error, isLoaded, movie } = this.state;

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
          className={'movie-banner-row'}
        >
          <div className={'movie-banner-row-filler'}>
            <Col>
              <Row className={'align-items-center'}>
                <Col sm={12} md={3}>
                  <img
                    className={'movie-poster'}
                    src={`https://image.tmdb.org/t/p/w500${movie.poster_path}`}
                  ></img>
                </Col>
                <Col sm={12} md={6} className={'movie-details'}>
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
        <Row className={'justify-content-center'}>
          <TorrentStream movie={movie} />
        </Row>
        <br />
        <br />
      </Container>
    );
  }
}
