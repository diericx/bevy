import React from 'react';
import Row from 'react-bootstrap/Row';
import Container from 'react-bootstrap/Container';
import Col from 'react-bootstrap/Col';
import './movie.css';
import TorrentStream from '../components/TorrentStream';
import { TmdbAPI } from '../lib/IceetimeAPI';

export default class MyComponent extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      movie: null,
      torrent: null,
    };
  }

  async componentDidMount() {
    let {
      location: {
        state: { movie },
      },
    } = this.props;
    if (!movie.externalIDs) {
      const resp = await TmdbAPI.GetMovie(movie.id);
      this.setState({
        isLoaded: true,
        movie: {
          ...resp,
          imdb_id: resp.imdb_id,
        },
      });
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
            backgroundImage: `url("${movie.backdrop_url}")`,
          }}
          className={'movie-banner-row'}
        >
          <div className={'movie-banner-row-filler'}>
            <Col>
              <Row className={'align-items-center'}>
                <Col sm={12} md={3}>
                  <img
                    className={'movie-poster'}
                    src={`${movie.poster_url}`}
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
