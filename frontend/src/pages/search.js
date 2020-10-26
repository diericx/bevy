import React from 'react';
import { Redirect, Link } from 'react-router-dom';
import Button from 'react-bootstrap/Button';
import VideoPlayer from '../components/VideoPlayer';
import Container from 'react-bootstrap/Container';
import Row from 'react-bootstrap/Row';
import Col from 'react-bootstrap/Col';
import Card from 'react-bootstrap/Card';
import './search.css';
import { TmdbAPI } from '../lib/IceetimeAPI';

export default class MyComponent extends React.Component {
  state = {
    query: null,
    resp: null,
  };

  async componentDidMount() {
    let {
      location: {
        state: { query },
      },
    } = this.props;
    const resp = await TmdbAPI.SearchMovie(query);
    console.log(resp);
    this.setState({
      isLoaded: true,
      resp,
    });
  }

  onMovieClick(movie) {
    this.setState({ redirect: { to: '/movie', state: { movie } } });
  }

  render() {
    const { error, isLoaded, resp, redirect } = this.state;
    if (error) {
      return <div>Error: {error.message}</div>;
    } else if (!isLoaded) {
      return <div>Loading...</div>;
    }

    if (redirect) {
      return (
        <Redirect
          push
          to={{
            pathname: redirect.to,
            state: redirect.state,
          }}
        />
      );
    }

    return (
      <Container>
        <br />
        {resp.results.map((item) => (
          <Row className="movie-row">
            <Col>
              <Card
                className="movie-card"
                onClick={() => this.onMovieClick(item)}
              >
                <Row noGutters>
                  <Col xs={1} className="movie-card-poster-col">
                    <Link
                      to={{
                        pathname: '/movie',
                        state: { movie: item },
                      }}
                    >
                      {!item.poster_url ? (
                        <Card.Img variant="top" className="movie-card-img" />
                      ) : (
                        <Card.Img
                          variant="top"
                          className="movie-card-img"
                          src={`${item.poster_url}`}
                        />
                      )}
                    </Link>
                  </Col>

                  <Col
                    xs={11}
                    className="align-items-center"
                    style={{ textAlign: 'left' }}
                  >
                    <Card.Body>
                      <Card.Title>
                        <b>{item.title}</b> {item.vote_average}
                      </Card.Title>
                      <Card.Subtitle className={'card-subtitle'}>
                        {item.overview}
                      </Card.Subtitle>
                    </Card.Body>
                  </Col>
                </Row>
              </Card>
            </Col>
          </Row>
        ))}
      </Container>
    );
  }
}
