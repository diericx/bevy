import React from 'react';
import { Redirect } from 'react-router-dom';

export default class MyComponent extends React.Component {
    constructor(props) {
      super(props);
      this.state = {
        movie: null,
      };
    }
  
    componentDidMount() {
      let { location: { state: { movie } } } = this.props;
      if (!movie.externalIDs) {
        let tmdbAPIKey = process.env.REACT_APP_TMDB_API_KEY;
        fetch(`https://api.themoviedb.org/3/movie/${movie.id}/external_ids?api_key=${tmdbAPIKey}`)
          .then(res => res.json())
          .then(
            (result) => {
              this.setState({
                isLoaded: true,
                movie: {
                  ...movie,
                  externalIDs: result
                }
              });
            },
            (error) => {
              this.setState({
                isLoaded: true,
                error
              });
            }
          )
      } else {
        this.setState({
          isLoaded: true,
          movie: movie,
        });
      }
    }
  
    render() {
      const { error, isLoaded, movie } = this.state;
      console.log(movie)
      if (error) {
        return <div>Error: {error.message}</div>;
      } else if (!isLoaded) {
        return <div>Loading...</div>;
      } else {
        return (
          <ul>
            Movie!
          </ul>
        );
      }
    }
}