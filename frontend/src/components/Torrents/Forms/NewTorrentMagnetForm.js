import React from "react";
import Form from "react-bootstrap/Form";
import Button from "react-bootstrap/Button";

export default class NewTorrentMagnetForm extends React.Component {
  state = {
    magnet: "",
  };

  handleChange = (e) => {
    this.setState({ [e.target.name]: e.target.value });
  };

  handleSubmit = (event) => {
    event.preventDefault();
    console.log("add torrent", this.state.magnet);
  };

  render() {
    let { magnet } = this.state;
    return (
      <Form onSubmit={this.handleSubmit}>
        <Form.Group controlId="formBasicEmail">
          <Form.Label>Magnet Link</Form.Label>
          <Form.Control
            name="magnet"
            placeholder="Enter magnet"
            value={magnet}
            onChange={this.handleChange}
          />
        </Form.Group>

        <Button variant="primary" type="submit">
          Submit
        </Button>
      </Form>
    );
  }
}
