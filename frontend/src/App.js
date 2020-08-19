import React from 'react';
import logo from './logo.svg';
import Navbar from './components/Navbar';
import Movies from './pages/movies';
import Movie from './pages/movie';
import './App.css';
import 'bootstrap/dist/css/bootstrap.min.css';
import 'video.js/dist/video-js.css';
import {
  BrowserRouter as Router,
  Switch,
  Route,
  Link
} from "react-router-dom";

require('dotenv').config()

export default function App() {
  return (
    <Router>
      <div>
        <Navbar />

        {/* A <Switch> looks through its children <Route>s and
            renders the first one that matches the current URL. */}
        <Switch>
          <Route path="/movie" render={(props) => <Movie {...props}/>}/>
          <Route path="/" render={(props) => <Movies {...props}/>}/>
        </Switch>
      </div>
    </Router>
  );
}
