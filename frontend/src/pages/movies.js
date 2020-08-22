import React from 'react';
import { Redirect, Link } from 'react-router-dom';
import Container from 'react-bootstrap/Container';
import Row from 'react-bootstrap/Row';
import Col from 'react-bootstrap/Col';
import Card from 'react-bootstrap/Card';
import Button from 'react-bootstrap/Button';

const styles = {
  movieCard: {
    paddingTop: "1em",
    paddingBottom: "1em",
  }
};

export default class MyComponent extends React.Component {
    constructor(props) {
      super(props);
      this.state = {
        error: null,
        isLoaded: false,
        resp: null,
      };
    }

    componentDidMount() {
      let tmdbAPIKey = process.env.REACT_APP_TMDB_API_KEY;
      fetch(`https://api.themoviedb.org/3/movie/popular?api_key=${tmdbAPIKey}`)
        .then(res => res.json())
        .then(
          (result) => {
            this.setState({
              isLoaded: true,
              resp: result
            });
          },
          (error) => {
            this.setState({
              isLoaded: true,
              error
            });
          }
        )
    }

    render() {
      const { error, isLoaded, resp, selectedMovie } = this.state;
      if (error) {
        return <div>Error: {error.message}</div>;
      } else if (!isLoaded) {
        return <div>Loading...</div>;
      } else {
        return (
          <Container>
            <Row className="justify-content-center">
              <Col sm={12}>
                <h1>Popular</h1>
                <hr></hr>
              </Col>

              {resp.results.map(item => (
                <Link
                  to={{
                    pathname: "/movie",
                    state: { movie: item}
                  }}
                  style={styles.movieCard}
                >
                  <Col>
                    <Card style={{ width: '15rem' }}>
                      <Card.Img variant="top" src={`https://image.tmdb.org/t/p/w500${item.poster_path}`} />
                      <Card.Body>
                        <Card.Title>{item.title} {item.vote_average}</Card.Title>
                      </Card.Body>
                    </Card>
                  </Col>
                </Link>
              ))}
            </Row>
          </Container>
          // <ul>
          //   {resp.results.map(item => (
          //     <Link
          //       to={{
          //         pathname: "/movie",
          //         state: { movie: item}
          //       }}
          //     >
          //       <li key={item.name}>
          //         {item.title} {item.vote_average}
          //       </li>
          //     </Link>
          //   ))}
          // </ul>
        );
      }
    }
}
