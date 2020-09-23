import React, { useState } from 'react';
import Button from 'react-bootstrap/Button';
import Modal from 'react-bootstrap/Modal';
import NewTorrentMagnetForm from './Forms/NewTorrentMagnetForm';

export default class NewTorrent extends React.Component {
  state = {
    show: false,
    magnet: null,
  };

  handleClose = () => this.setState({ show: false });
  handleShow = () => this.setState({ show: true });

  render() {
    let { show } = this.state;
    return (
      <>
        <Button variant="primary" onClick={this.handleShow}>
          New Torrent Magnet
        </Button>

        <Modal show={show} onHide={this.handleClose}>
          <Modal.Header closeButton>
            <Modal.Title>New Torrent Magnet</Modal.Title>
          </Modal.Header>
          <Modal.Body>
            <NewTorrentMagnetForm
              successHook={this.handleClose}
              formState={this.state}
              handleChange={this.handleChange}
            />
          </Modal.Body>
        </Modal>
      </>
    );
  }
}
