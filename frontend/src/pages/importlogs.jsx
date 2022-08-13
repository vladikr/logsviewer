import React, {useState} from 'react';
import axios from 'axios';
import {DashboardLayout} from '../components/Layout';

const ImportLogsPage = () => {
  const [file, setFile] = useState()
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
    axios.post(url, formData, config).then((response) => {
      console.log(response.data);
    }).catch(error => {
    console.log(error.response)
});

  }

  return (
    <DashboardLayout>
      <form onSubmit={handleSubmit}>
          <h1>Import Logs</h1>
          <input type="file" onChange={handleChange}/>
          <button type="submit">Upload</button>
        </form>
    </DashboardLayout>
  )
}

export default ImportLogsPage;
