import * as React from 'react';
import "@patternfly/react-core/dist/styles/base.css";
import axios from 'axios';
import { Migrations } from '@app/Workloads/Migrations/Migrations';
import { YAMLEditor } from "@app/Common/Editor"
import {
  Tabs, Tab, TabTitleText,
  TabContent,
  TabContentBody,
  Card,
  Pagination,
  PageSection,
  Flex,
  FlexItem,
  Toolbar,
  ToolbarContent,
  ToolbarGroup,
  ToolbarItem,
  Bullseye, EmptyState, EmptyStateIcon, Spinner, Title,
} from "@patternfly/react-core";

interface PodTabsProps {
    namespace?: string,
    name?: string,
    uuid: string
}

const PodTabs: React.FunctionComponent<PodTabsProps> = ({name, namespace, uuid}: PodTabsProps) => { 
	const [loadingYamlData, setLoadingYamlData] = React.useState(true);
    const [activeTabKey, setActiveTabKey] = React.useState<string | number>(0);
  	const [data, setData] = React.useState("Empty");
  	React.useEffect(() => {
    	async function getData(uuid: string) {
      	await axios
        	.get("/getPodYaml",
            {
                params: {
                    uuid: uuid,
                }
            })
        	.then((response) => {
          	// check if the data is populated
          	console.log(response.data);
          	console.log(response.data.yaml);
            if (response.data.yaml) {
                const processedData = response.data.yaml;
          	    setData(processedData);
          	    // you tell it that you had the result
          	    setLoadingYamlData(false);
            }
        	});
    	}
    if (loadingYamlData) {
      // if the result is not ready so you make the axios call
      getData(uuid);
    }
  }, []);

    const handleTabClick = (
        event: React.MouseEvent<any> | React.KeyboardEvent | MouseEvent,
        tabIndex: string | number
    ) => {
        setActiveTabKey(tabIndex);
    };
    
    const renderEditor = (yamlData: string) => {
        return (
            <YAMLEditor data={yamlData} />
    )}

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


        //usePageInsets
return (
    <div>
        <Tabs
        isFilled
        activeKey={activeTabKey}
        onSelect={handleTabClick}
        isBox={false}
        aria-label="Tabs in the page insets example"
        role="region"
      >
            <Tab eventKey={0} title={<TabTitleText>YAML</TabTitleText>} aria-label="Pods Yaml" tabContentId={`tabContent${0}`} />
            <Tab eventKey={1} title={<TabTitleText>Events</TabTitleText>} aria-label="Events" tabContentId={`tabContent${1}`} />
            <Tab eventKey={2} title={<TabTitleText>Storage</TabTitleText>} aria-label="Storage" tabContentId={`tabContent${2}`} />
            <Tab eventKey={3} title={<TabTitleText>Events</TabTitleText>} aria-label="Networking" tabContentId={`tabContent${3}`} />
        </Tabs>
        <TabContent
          key={0}
          eventKey={0}
          id={`tabContent${0}`}
          activeKey={activeTabKey}
          hidden={0 !== activeTabKey}
        >
          <TabContentBody>
            <Flex direction={{ default: 'column' }} spaceItems={{ default: 'spaceItemsLg' }}>
              <FlexItem>
                { loadingYamlData || 0 !== activeTabKey ? (loadingElem()) : (renderEditor(data))}
              </FlexItem>
            </Flex>
          </TabContentBody>
        </TabContent>
        <TabContent
          key={1}
          eventKey={1}
          id={`tabContent${1}`}
          activeKey={activeTabKey}
          hidden={1 !== activeTabKey}
        >
          <TabContentBody>
              <div>TBD</div>
          </TabContentBody>
        </TabContent>
        <TabContent
          key={2}
          eventKey={2}
          id={`tabContent${2}`}
          activeKey={activeTabKey}
          hidden={2 !== activeTabKey}
        >
          <TabContentBody>
              <div>TBD</div>
          </TabContentBody>
        </TabContent>
        <TabContent
          key={3}
          eventKey={3}
          id={`tabContent${3}`}
          activeKey={activeTabKey}
          hidden={3 !== activeTabKey}
        >
          <TabContentBody>
              <div>TBD</div>
          </TabContentBody>
        </TabContent>
    </div>
);}

export { PodTabs };
