import * as React from 'react';
import MonacoEditor from 'react-monaco-editor';
import Measure from 'react-measure';
import {
  PageSection,
} from "@patternfly/react-core";

import './YAMLEditor.scss'; 


interface EditorProps {
    data?: string
}

const YAMLEditor: React.FunctionComponent<EditorProps> = ({data}: EditorProps) => { 

    const onEditorDidMount = (editor, monaco) => {
        //eslint-disable-next-line no-console
        console.log("editor value: ", editor.getValue());
        editor.layout();
        editor.focus();
        monaco.editor.getModels()[0].updateOptions({ tabSize: 5 });
    };

    const onChange = value => {
        // eslint-disable-next-line no-console
        console.log(value);
    };
    console.log("got data: ", data);
return (
    <>
       <Measure bounds>
        {({ measureRef, contentRect }) => (
          <div ref={measureRef} className="ocs-yaml-editor__root" >
            <div className="ocs-yaml-editor__wrapper">

                <MonacoEditor
                   language="yaml"
                   theme="console"
                   height="80%"
                   width="400px"
                   value={data}
                   editorDidMount={onEditorDidMount}
                   onChange={onChange}
                />
            </div>
          </div>
        )}
      </Measure>
    </>

);}

export { YAMLEditor };
