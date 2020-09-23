import React from 'react';
import { Redirect, Link } from 'react-router-dom';
import Container from 'react-bootstrap/Container';
import Row from 'react-bootstrap/Row';
import Col from 'react-bootstrap/Col';
import Card from 'react-bootstrap/Card';
import Form from 'react-bootstrap/Form';

const styles = {
  movieCard: {
    width: '15rem',
    boxShadow: '0px 0px 10px gray',
    border: 'none',
    marginTop: '1em',
    marginBottom: '1em',
  },
};

let tmdbAPIKey = window._env_.REACT_APP_TMDB_API_KEY;

export default class MyComponent extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      error: null,
      isLoaded: false,
      resp: null,
      searchQuery: '',
      toSearch: false,
    };
  }

  handleSearchFormSubmit = (event) => {
    event.preventDefault();
    const { searchQuery } = this.state;

    if (searchQuery == '') {
      return;
    }

    this.setState({
      toSearch: true,
    });
  };

  handleSearchQueryChange = (event) => {
    this.setState({ searchQuery: event.target.value });
  };

  componentDidMount() {
    fetch(`https://api.themoviedb.org/3/movie/popular?api_key=${tmdbAPIKey}`)
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
    const {
      error,
      isLoaded,
      resp,
      toSearch,
      searchQuery,
      selectedMovie,
    } = this.state;
    if (error) {
      return <div>Error: {error.message}</div>;
    } else if (!isLoaded) {
      return <div>Loading...</div>;
    }

    if (toSearch) {
      return (
        <Redirect
          push
          to={{
            pathname: '/search',
            state: { query: searchQuery },
          }}
        />
      );
    }

    return (
      <Container>
        <Row>
          <Col sm={12}>
            <h1>Search</h1>
            <hr></hr>
          </Col>
          <Col>
            <Form onSubmit={this.handleSearchFormSubmit}>
              <Form.Group controlId="formMovieSearch">
                <Form.Control
                  type="search"
                  placeholder="Enter search"
                  value={this.state.searchQuery}
                  onChange={this.handleSearchQueryChange}
                />
              </Form.Group>
            </Form>
          </Col>
        </Row>

        <Row className="justify-content-center">
          <Col sm={12}>
            <h1>Popular</h1>
            <hr></hr>
          </Col>

          {resp.results.map((item) => (
            <Col>
              <Card style={styles.movieCard}>
                <Link
                  to={{
                    pathname: '/movie',
                    state: { movie: item },
                  }}
                >
                  <Card.Img
                    variant="top"
                    src={`https://image.tmdb.org/t/p/w500${item.poster_path}`}
                  />
                </Link>
                <Card.Body>
                  <Card.Title>
                    {item.title} {item.vote_average}
                  </Card.Title>
                </Card.Body>
              </Card>
            </Col>
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
