import React from "react";
import { Redirect, Link } from "react-router-dom";
import Button from "react-bootstrap/Button";
import VideoPlayer from "../components/VideoPlayer";
import Container from "react-bootstrap/Container";
import Row from "react-bootstrap/Row";
import Col from "react-bootstrap/Col";
import Card from "react-bootstrap/Card";
import "./search.css";

export default class MyComponent extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      query: null,
      response: null,
    };
  }

  componentDidMount() {
    let {
      location: {
        state: { query },
      },
    } = this.props;
    let tmdbAPIKey = process.env.REACT_APP_TMDB_API_KEY;
    console.log("Got to search with: ", query);
    fetch(
      `https://api.themoviedb.org/3/search/movie?api_key=${tmdbAPIKey}&query=${query}`
    )
      .then((res) => res.json())
      .then(
        (result) => {
          this.setState({
            isLoaded: true,
            resp: result,
          });
        },
        (error) => {
          this.setState({
            isLoaded: true,
            error,
          });
        }
      );
  }

  onMovieClick(movie) {
    this.setState({redirect: {to: '/movie', state: {movie} }})
  }

  render() {
    const { error, isLoaded, resp, redirect } = this.state;
    if (error) {
      return <div>Error: {error.message}</div>;
    } else if (!isLoaded) {
      return <div>Loading...</div>;
    }

    if (redirect) {
      return <Redirect
        push
        to={{
          pathname: redirect.to,
          state: redirect.state,
        }}
      />
    }

    return (
      <Container>
        {resp.results.map((item) => (
          <Row className="movie-row">
            <Col>
              <Card className="movie-card" onClick={() => this.onMovieClick(item)}>
                <Row noGutters>
                  <Col xs={1} className="movie-card-poster-col">
                    <Link
                      to={{
                        pathname: "/movie",
                        state: { movie: item },
                      }}
                    >
                      <Card.Img
                        variant="top"
                        className="movie-card-img"
                        src={`https://image.tmdb.org/t/p/w500${item.poster_path}`}
                      />
                    </Link>
                  </Col>

                  <Col xs={11} className="align-items-center" style={{ textAlign: "left" }}>
                    <Card.Body>
                      <Card.Title>
                        <b>{item.title}</b> {item.vote_average}
                      </Card.Title>
                      <Card.Subtitle className={"card-subtitle"}>
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
