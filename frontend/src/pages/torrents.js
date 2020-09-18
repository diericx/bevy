import React from "react";
import PropTypes from "prop-types";
import Row from "react-bootstrap/Row";
import Container from "react-bootstrap/Container";
import Col from "react-bootstrap/Col";
import Table from "react-bootstrap/Table";
import Spinner from "react-bootstrap/Table";
import "./movie.css";
import TorrentStream from "../components/TorrentStream";
import { result } from "underscore";
import NewTorrent from "../components/Torrents/NewTorrent.js";
import { TorrentsAPI } from "../lib/IceetimeAPI";

let backendURL = window._env_.BACKEND_URL;

export default class Torrents extends React.Component {
  state = {
    torrents: null,
    error: null,
    isLoaded: false,
  };

  async componentDidMount() {
    this.setState({
      isLoaded: true,
      ...(await TorrentsAPI.Get()),
    });
  }

  render() {
    let { torrents, isLoaded, error } = this.state;
    if (!isLoaded) {
      return <Spinner animation="border" role="status"></Spinner>;
    }
    if (error) {
      return <p>{error.toString()}</p>;
    }

    return (
      <Container>
        <Row>
          <NewTorrent />
        </Row>
        <Row>
          <Col xs={12}>
            <Table striped bordered hover size="sm">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Size</th>
                  <th>Progress</th>
                </tr>
              </thead>
              <tbody>
                {torrents.map((torrent) => (
                  <tr>
                    <td>{torrent.name}</td>
                    <td>{torrent.length}</td>
                    <td>0</td>
                  </tr>
                ))}
              </tbody>
            </Table>
          </Col>
        </Row>
      </Container>
    );
  }
}
