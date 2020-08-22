import React from 'react';
import { Redirect, Link } from 'react-router-dom';

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
      let tmdbAPIKey = window._env_.REACT_APP_TMDB_API_KEY;
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
          <ul>
            {resp.results.map(item => (
              <Link
                to={{
                  pathname: "/movie",
                  state: { movie: item}
                }}
              >
                <li key={item.name}>
                  {item.title} {item.vote_average}
                </li>
              </Link>
            ))}
          </ul>
        );
      }
    }
}
