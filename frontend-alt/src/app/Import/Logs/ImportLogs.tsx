import * as React from 'react';
import axios from 'axios';
import { PageSection, Title, FileUpload, Bullseye, Card, EmptyState, EmptyStateIcon, Spinner } from '@patternfly/react-core';
import "@patternfly/react-core/dist/styles/base.css";



const ImportLogs: React.FunctionComponent = () => {
  const [filename, setFilename] = React.useState('');
  const [isLoading, setIsLoading] = React.useState(false);
  const [errorMessage, setErrorMessage] = React.useState("");


const handleFileInputChange = (
    _event: React.ChangeEvent<HTMLInputElement> | React.DragEvent<HTMLElement>,
    file: File
  ) => {
    setErrorMessage('');
    setFilename(file.name);
    setIsLoading(true);
    setErrorMessage(`Uploading.. ${file.name} - ${isLoading}`);
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
        setErrorMessage('');
        setIsLoading(false);
    }).catch(error => {
        setErrorMessage(`Unable to load logs: ${error.response}`);
        console.log(error.response)
        setIsLoading(false);
    });
    }

  const handleClear = (_event: React.MouseEvent<HTMLButtonElement, MouseEvent>) => {
    setFilename('');
    setErrorMessage('');
  };
  const loadingElem = () => (
                <EmptyState>
                  <EmptyStateIcon variant="container" component={Spinner} />
                  <Title size="lg" headingLevel="h2">
                    Uploading {filename}
                  </Title>
                </EmptyState>
  )
const uploadForm = () => {
             const isLoadingUpdate = isLoading
             return (
                
                <FileUpload
                  id="fileUploadForm"
                  filename={filename}
                  filenamePlaceholder="Drag and drop a file or upload one"
                  onFileInputChange={handleFileInputChange}
                  onClearClick={handleClear}
                  browseButtonText="Upload"
                >
                {errorMessage.length !== 0 && <div className="pf-u-m-md">{errorMessage}</div>}
                </FileUpload>
)}

const isLoadingUpdate = isLoading;
return (
      <PageSection>
        <Bullseye>
              { isLoadingUpdate ? (loadingElem()) : (uploadForm())}
        </Bullseye>
      </PageSection>
    );
}

export { ImportLogs };
