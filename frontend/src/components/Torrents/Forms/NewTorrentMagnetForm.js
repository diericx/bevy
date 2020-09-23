import React from 'react';
import Form from 'react-bootstrap/Form';
import Button from 'react-bootstrap/Button';
import Alert from 'react-bootstrap/Alert';
import { TorrentsAPI } from '../../../lib/IceetimeAPI';
import Spinner from 'react-bootstrap/esm/Spinner';

export default class NewTorrentMagnetForm extends React.Component {
  state = {
    magnet: '',
    isLoaded: true,
  };

  handleChange = (e) => {
    this.setState({ [e.target.name]: e.target.value });
  };

  handleSubmit = async (event) => {
    const { magnet } = this.state;
    const { successHook } = this.props;
    event.preventDefault();

    this.setState({ isLoaded: false });
    const resp = await TorrentsAPI.NewMagnet(magnet);
    this.setState({
      isLoaded: true,
      ...resp,
    });

    if (!resp.error) {
      successHook();
    }
  };

  render() {
    let { magnet, error, isLoaded } = this.state;
    return (
      <>
        {error ? <Alert variant={'danger'}>{error.toString()}</Alert> : null}
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

          {isLoaded ? (
            <Button variant="primary" type="submit">
              Submit
            </Button>
          ) : (
            <Spinner animation="border" role="status"></Spinner>
          )}
        </Form>
      </>
    );
  }
}
