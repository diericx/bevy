import React from "react";
import PropTypes from "prop-types";
import Row from "react-bootstrap/Row";
import Container from "react-bootstrap/Container";
import Col from "react-bootstrap/Col";
import Table from "react-bootstrap/Table";
import ProgressBar from "react-bootstrap/ProgressBar";
import Spinner from "react-bootstrap/Table";
import "./movie.css";
import prettyBytes from "pretty-bytes";
import TorrentStream from "../components/TorrentStream";
import { result } from "underscore";
import NewTorrent from "../components/Torrents/NewTorrent.js";
import { TorrentsAPI } from "../lib/IceetimeAPI";

let backendURL = window._env_.BACKEND_URL;
const REFRESH_RATE = 2000;

export default class Torrents extends React.Component {
  state = {
    torrents: null,
    error: null,
    isLoaded: false,
  };

  componentDidMount() {
    this.fetchData();
    this.timer = setInterval(() => this.fetchData(), REFRESH_RATE);
  }
  componentWillUnmount() {
    clearInterval(this.timer);
  }

  async fetchData() {
    const resp = await TorrentsAPI.Get();
    this.setState({
      isLoaded: true,
      ...resp,
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
                  <th>Peers</th>
                  <th>Uploaded</th>
                  <th>Downloaded</th>
                </tr>
              </thead>
              <tbody>
                {torrents.map((torrent) => {
                  const progress =
                    (100 * torrent.bytesCompleted) / torrent.length;
                  return (
                    <tr>
                      <td>{torrent.name}</td>
                      <td>{prettyBytes(torrent.length)}</td>
                      <td>
                        <ProgressBar
                          now={progress}
                          label={`${Math.round(progress)}%`}
                        />
                      </td>
                      <td>
                        {torrent.activePeers}({torrent.totalPeers})
                      </td>
                      <td>{prettyBytes(torrent.bytesWrittenData)}</td>
                      <td>{prettyBytes(torrent.bytesReadData)}</td>
                    </tr>
                  );
                })}
              </tbody>
            </Table>
          </Col>
        </Row>
      </Container>
    );
  }
}
