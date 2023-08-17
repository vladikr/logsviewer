import * as React from 'react';
import "@patternfly/react-core/dist/styles/base.css";
import axios from 'axios';

import {
  Button,
  DescriptionList,
  DescriptionListTerm,
  DescriptionListGroup,
  DescriptionListDescription,
  Bullseye, EmptyState, EmptyStateIcon, Spinner, Title,
} from '@patternfly/react-core';
import PlusCircleIcon from '@patternfly/react-icons/dist/esm/icons/plus-circle-icon';

import { apiBaseUrl } from "@app/config";

interface VMIDetailsProps {
    uuid?: string,
    nodeName?: string,
}

  const columnNames = {
    podName: "Pod Name",
    podUUID: "Pod UUID",
    virtHandlerName: "Virt-Handler Pod Name",
    creationTime: "Creation Time",
    pvcs: "PVCs",
  };

    type VmiDetails = {
        podName: string;
        podUUID: string;
        virtHandlerName: string;
        creationTime: Date;
        pvcs: string[];
    };

const VMIDetailsMinimal: React.FunctionComponent<VMIDetailsProps> = ({uuid, nodeName}: VMIDetailsProps) => {

	const [loadingData, setLoadingData] = React.useState(true);
  	const [data, setData] = React.useState<VmiDetails>({} as VmiDetails);
    const vmiDetails: VmiDetails = data;

  	React.useEffect(() => {
        //let apiVerb = renderAPIGet(object)
        let apiVerb = '/getVMIDetails'
    	async function getData() {
      	await axios
        	.get(apiBaseUrl + apiVerb!,
            {
                params: {
                    uuid: uuid,
                    nodeName: nodeName
                }
            })
        	.then((response) => {
          	// check if the data is populated
          	console.log(response.data);
            const res = response.data;
            // convert data to VmiDetails
            const processedData: VmiDetails = { 
                    podName: res.SourcePod, 
                    podUUID: res.SourcePodUUID,
                    virtHandlerName: res.SourceHandler,
                    creationTime: new Date(res.StartTimestamp),
                    pvcs: (res.PVCs ?? [])
            };
          	setData(processedData);
          	setLoadingData(false);
        	});
    	}
        if (loadingData) {
          // if the result is not ready so you make the axios call
          getData();
        }
      }, []);

    const generateVMIDetailsFields = (data) => (
        Object.keys(columnNames).map((key, index) => {
            return (
                <DescriptionListGroup>
                    <DescriptionListTerm>{columnNames[key]}</DescriptionListTerm>
                    <DescriptionListDescription>{data[key]?.toString()}</DescriptionListDescription>
                </DescriptionListGroup>
            );
        }))

    const renderVMIDetails = () => {
        return (
          <DescriptionList isAutoColumnWidths columnModifier={{ default: '3Col' }}>
            {generateVMIDetailsFields(data)}
          </DescriptionList>
        ) 
    }

    const loadingElem = () => (
        <Bullseye>
                <EmptyState>
                  <EmptyStateIcon variant="container" component={Spinner} />
                  <Title size="lg" headingLevel="h2">
                    Loading
                  </Title>
                </EmptyState>
              </Bullseye>
  )

    return ( loadingData ? (loadingElem()) : (renderVMIDetails()) )
}

export { VMIDetailsMinimal };
