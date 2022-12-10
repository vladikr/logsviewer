import React, {useState} from 'react';
import axios from 'axios';
import {DashboardLayout} from '../components/Layout';
import {LoadingSpinner} from '../components/Spinner';
import Modal from "react-bootstrap/Modal";
import Button from 'react-bootstrap/Button';
import Form from 'react-bootstrap/Form';

const ImportLogsPage = () => {
  const [file, setFile] = useState()
  const [isLoading, setIsLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");
  const [show, setShow] = useState(true);

  const handleClose = () => setShow(false);
  const handleShow = () => setShow(true);

  function handleChange(event) {
    setFile(event.target.files[0])
  }
  function handleSubmit(event) {
    event.preventDefault()
    //const url = 'http://localhost:8080/uploadLogs';
    const url = '/uploadLogs';
    const formData = new FormData();
    formData.append('file', file);
    formData.append('fileName', file.name);
    const config = {
      headers: {
        'content-type': 'multipart/form-data',
      },
    };
    setIsLoading(true);
    axios.post(url, formData, config).then((response) => {
      console.log(response.data);
      setIsLoading(false);
    }).catch(error => {
        setErrorMessage(`Unable to load logs: ${error.response}`);
        setIsLoading(false);
        console.log(error.response)
});

  }
  const uploadForm = (
      //<form id='upload-form' onSubmit={handleSubmit}>
      //    <input type="file" onChange={handleChange}/>
      //    <Button type="submit" variant="primary">Upload</Button>
      //  </form>
    <Form noValidate onSubmit={handleSubmit}>
        <Form.Group controlId="formFile" className="mb-3">
            <Form.Label>Upload compressed must-gather files </Form.Label>
            <Form.Control type="file" onChange={handleChange} />
        </Form.Group>

        <Button variant="primary" type="submit">
            Upload
        </Button>
    </Form>
    );
  return (
    <DashboardLayout>
        <Modal 
        size="lg"
        aria-labelledby="contained-modal-title-vcenter"
        centered
        show={show}
        onHide={handleClose}
      >
            <Modal.Header closeButton>
              <Modal.Title>Upload Logs</Modal.Title>
            </Modal.Header> 
            <Modal.Body>
                {isLoading ? <LoadingSpinner />: uploadForm}
                {errorMessage && <div className="error">{errorMessage}</div>}
            </Modal.Body>
      </Modal>
    </DashboardLayout>
  )
}

export default ImportLogsPage;
