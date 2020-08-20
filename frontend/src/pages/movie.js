import React from 'react';
import { Redirect } from 'react-router-dom';
import Button from 'react-bootstrap/Button'
import VideoPlayer from '../components/VideoPlayer';

export default class MyComponent extends React.Component {
    constructor(props) {
      super(props);
      this.state = {
        movie: null,
        torrent: null,
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

    findTorrent(imdbID, title, year) {
      fetch(`http://localhost:8080/find/movie?imdbid=${imdbID}&title=${title}&year=${year}`)
        .then(res => res.json())
        .then(
          (result) => {
            this.setState({
              isLoaded: true,
              torrent: result 
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
      const { error, isLoaded, movie, torrent } = this.state;

      if (error) {
        return <div>Error: {error.message}</div>;
      } else if (!isLoaded) {
        return <div>Loading...</div>;
      }

      if (!torrent) {
        let releaseDate = movie.release_date.split("-")[0]
        return (
          <Button variant="primary" onClick={() => this.findTorrent(movie.externalIDs.imdb_id, movie.title, releaseDate)}>Fetch Movie</Button>
        ) 
      }

      const videoJsOptions = {
        autoplay: true,
        controls: true,
        width: 720,
        plugins: {
          timeRangesSeeking: {},
          durationFromServer: {},
          videoJsResolutionSwitcher: {
            default: 'high',
            dynamicLabel: true,
          },
        },
        sources: [
          {
            src: `http://localhost:8080/stream/torrent/${torrent.id}/transcode`,
            type: 'video/mp4',
            label: '720p',
          },
          {
            src: `http://localhost:8080/stream/torrent/${torrent.id}/transcode?resolution=1080p`,
            type: 'video/mp4',
            label: '1080p',
          }
        ]
      }

      return (
        <div>
          <VideoPlayer {...videoJsOptions} />
        </div>
      );
    }
}