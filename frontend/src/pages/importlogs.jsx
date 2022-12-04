import React, {useState} from 'react';
import axios from 'axios';
import {DashboardLayout} from '../components/Layout';
import {LoadingSpinner} from '../components/Spinner';

const ImportLogsPage = () => {
  const [file, setFile] = useState()
  const [isLoading, setIsLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");

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
      <form onSubmit={handleSubmit}>
          <h1>Import Logs</h1>
          <input type="file" onChange={handleChange}/>
          <button type="submit">Upload</button>
        </form>
    );
  return (
    <DashboardLayout>
        {isLoading ? <LoadingSpinner />: uploadForm}
        {errorMessage && <div className="error">{errorMessage}</div>}
    </DashboardLayout>
  )
}

export default ImportLogsPage;
