import React from 'react';
import Row from 'react-bootstrap/Row';
import Container from 'react-bootstrap/Container';
import Col from 'react-bootstrap/Col';
import Table from 'react-bootstrap/Table';
import Button from 'react-bootstrap/Button';
import './movie.css';
import TorrentStream from '../components/TorrentStream';
import { TmdbAPI } from '../lib/IceetimeAPI';
import { TorrentsAPI, TranscoderAPI } from '../lib/IceetimeAPI';
import Spinner from 'react-bootstrap/esm/Spinner';

export default class MyComponent extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      movie: null,
      torrent: null,
    };
  }

  async findTorrent(imdbID, title, year) {
    const resp = await TorrentsAPI.FindTorrentForMovie(imdbID, title, year, 0);
    this.setState({
      isFindTorrentCallLoading: false,
      ...resp,
    });
  }

  async scoredReleasesForMovie(imdbID, title, year) {
    const resp = await TorrentsAPI.ScoredReleasesForMovie(
      imdbID,
      title,
      year,
      0
    );
    console.log(resp);
    this.setState({
      isScoredReleasesCallLoading: false,
      ...resp,
    });
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

  FindTorrentButton = () => {
    const { torrentLink, isFindTorrentCallLoading, movie } = this.state;
    if (torrentLink) {
      return null;
    }
    if (isFindTorrentCallLoading) {
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
    return (
      <Button
        variant="primary"
        onClick={async () => {
          this.findTorrent(movie.imdb_id, movie.title, movie.release_year);
          this.setState({ isFindTorrentCallLoading: true });
        }}
      >
        Find Movie Automatically
      </Button>
    );
  };

  ManualSearchButton = () => {
    const { isScoredReleasesCallLoading, movie } = this.state;
    if (isScoredReleasesCallLoading) {
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
    return (
      <Button
        variant="primary"
        onClick={async () => {
          this.scoredReleasesForMovie(
            movie.imdb_id,
            movie.title,
            movie.release_year
          );
          this.setState({ isScoredReleasesCallLoading: true });
        }}
      >
        Manually Search for Movie
      </Button>
    );
  };

  render() {
    const {
      error,
      isLoaded,
      isFindLoading,
      movie,
      torrentLink,
      scoredReleases,
    } = this.state;

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

        <Row
          className={'justify-content-center'}
          style={{ textAlign: 'center' }}
        >
          <this.FindTorrentButton />
          <TorrentStream movie={movie} torrentLink={torrentLink} />
        </Row>
        <br />
        <Row
          className={'justify-content-center'}
          style={{ textAlign: 'center' }}
        >
          <this.ManualSearchButton />
          <ScoredReleases scoredReleases={scoredReleases} />
        </Row>
        <br />
        <br />
      </Container>
    );
  }
}

class ScoredReleases extends React.Component {
  render() {
    const { scoredReleases } = this.props;
    if (!scoredReleases) {
      return null;
    }

    return (
      <Table>
        <thead>
          <tr>
            <th>Title</th>
            <th>Size</th>
            <th>Seeders</th>
            <th>Size Score</th>
            <th>Seeder Score</th>
            <th>Quality Score</th>
            <th>Total Score</th>
          </tr>
        </thead>
        <tbody>
          {scoredReleases.map((scoredRelease) => (
            <tr>
              <td>{scoredRelease.Title}</td>
              <td>{scoredRelease.Size}</td>
              <td>{scoredRelease.Seeders}</td>
              <td>{scoredRelease.SizeScore.toFixed(2)}</td>
              <td>{scoredRelease.SeederScore.toFixed(2)}</td>
              <td>{scoredRelease.QualityScore.toFixed(2)}</td>
              <td>
                {(
                  scoredRelease.SeederScore +
                  scoredRelease.SizeScore +
                  scoredRelease.QualityScore
                ).toFixed(2)}
              </td>
            </tr>
          ))}
        </tbody>
      </Table>
    );
  }
}
