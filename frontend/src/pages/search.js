import React from "react";
import { Redirect, Link } from "react-router-dom";
import Button from "react-bootstrap/Button";
import VideoPlayer from "../components/VideoPlayer";
import Container from "react-bootstrap/Container";
import Row from "react-bootstrap/Row";
import Col from "react-bootstrap/Col";
import Card from "react-bootstrap/Card";

const styles = {
  movieCardPosterCol: {
    maxHeight: "8em",
  },

  moviePoster: {
    height: "100%",
    width: "4em",
    // maxWidth: "50px",
  },
};
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

  render() {
    const { error, isLoaded, resp } = this.state;
    if (error) {
      return <div>Error: {error.message}</div>;
    } else if (!isLoaded) {
      return <div>Loading...</div>;
    }

    return (
      <Container>
        {resp.results.map((item) => (
          <Row>
            <Col>
              <Card>
                <Row noGutters>
                  <Col md={1} style={styles.movieCardPosterCol}>
                    <Link
                      to={{
                        pathname: "/movie",
                        state: { movie: item },
                      }}
                    >
                      <Card.Img
                        variant="top"
                        style={styles.moviePoster}
                        src={`https://image.tmdb.org/t/p/w500${item.poster_path}`}
                      />
                    </Link>
                  </Col>

                  <Col md={11} style={{ textAlign: "left" }}>
                    <Card.Body>
                      <Card.Title>
                        {item.title} {item.vote_average}
                      </Card.Title>
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
