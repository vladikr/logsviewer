import * as React from 'react';
import "@patternfly/react-core/dist/styles/base.css";
import { PodTabs } from '@app/Workloads/Pods/PodTabs';
import axios from 'axios';
import {
  TableComposable,
  Thead,
  Tr,
  Th,
  Tbody,
  Td,
  ExpandableRowContent,
  ActionsColumn,
  IAction
} from "@patternfly/react-table";
import {
  Divider,
  Drawer,
  DrawerContent,
  DrawerContentBody,
  DrawerPanelContent,
  DrawerHead,
  DrawerActions,
  DrawerCloseButton,
  DrawerPanelBody,
  Flex,
  FlexItem,  
  Card,
  Pagination,
  PageSection,
  Tabs,
  Tab,
  TabContent,
  TabContentBody,
  TabTitleText,
  Toolbar,
  ToolbarContent,
  ToolbarGroup,
  ToolbarItem,
  Bullseye, EmptyState, EmptyStateIcon, Spinner, Title,
} from "@patternfly/react-core";
import { apiBaseUrl } from "@app/config";
import {useLocation} from "react-router-dom";
import * as queryString from "querystring";

const Pods: React.FunctionComponent = () => {
  const { search } = useLocation()
  const queryStringValues = queryString.parse(search.slice(1))

	const [loadingData, setLoadingData] = React.useState(true);
  	const [data, setData] = React.useState<any[]>([]);


  	React.useEffect(() => {
      let url = apiBaseUrl + "/pods";

      if (queryStringValues.status !== undefined) {
        url = url + "?status=" + queryStringValues.status
      }

    	async function getData() {
      	await axios
        	.get(url)
        	.then((response) => {
          	// check if the data is populated
          	console.log(response.data);
            const processedData = generatePodData(response.data.data);
          	setData(processedData);
            const localPods = processedData
            setPaginatedRows(localPods.slice(0, 10));
          	// you tell it that you had the result
          	setLoadingData(false);
        	});
    	}
    if (loadingData) {
      // if the result is not ready so you make the axios call
      getData();
    }
  }, []);

  type Pod = {
    uuid: string;
    name: string;
    namespace: string;
    phase: string;
    activeContainers: number;
    totalContainers: number;
    creationTime: Date;
    createdBy: string;
    nestedComponent?: React.ReactNode;
    link?: React.ReactNode;
    noPadding?: boolean;
  };
  const generatePodData = (unproccessedData: any[]) => {
    const pods: Pod[] = [];
    unproccessedData.map((res) => {
      //res['cretionTime'] = new Date(res.creationTime);
      const newRes: Pod = { ...res, creationTime: new Date(res.creationTime) };
      pods.push(newRes);
      return pods;
    });
    console.log(pods);
    return pods;
  };

  const pods: Pod[] = data;

  const [page, setPage] = React.useState(1);
  const [perPage, setPerPage] = React.useState(10);
  const [paginatedRows, setPaginatedRows] = React.useState(pods.slice(0, 10));
  const handleSetPage = (_evt, newPage, perPage, startIdx, endIdx) => {
    setPaginatedRows(pods.slice(startIdx, endIdx));
    setPage(newPage);
  };
  const handlePerPageSelect = (_evt, newPerPage, newPage, startIdx, endIdx) => {
    setPaginatedRows(pods.slice(startIdx, endIdx));
    setPage(newPage);
    setPerPage(newPerPage);
  };
  const fetchDSLQuery = async (
  	uuid: string
  ) => {
      const retq = await axios.get(apiBaseUrl + "/getSinglePodQueryParams",
          {
              params: {
                  uuid: uuid
              }
          }).then(function (resp) {
              console.log("await2: ", resp.data.dslQuery)
              const hostname = window.location.hostname
              const hostnameParts = hostname.split('.');
              const ingress = hostnameParts.slice(1).join('.');
              const appNameParts = hostnameParts.slice(0, 1)[0].split('-');
              let suffix = ""
              if (appNameParts.length > 1) {
                  suffix = "-" + appNameParts.slice(1);
              }
              const kibanaHostname = "kibana" + suffix + "." + ingress;
              
              window.open(`http://${kibanaHostname}/app/discover#/?${resp.data.dslQuery}`, '_blank', 'noopener,noreferrer');
              return {
                  query: resp.data.dslQuery, 
              };
          })
          
          return retq
  }

  const renderPagination = (variant, isCompact) => (
    <Pagination
      isCompact={isCompact}
      itemCount={pods.length}
      page={page}
      perPage={perPage}
      onSetPage={handleSetPage}
      onPerPageSelect={handlePerPageSelect}
      variant={variant}
      titles={{
        paginationTitle: `${variant} pagination`
      }}
    />
  );




  const columnNames = {
    uuid: "UUID",
    name: "Name",
    namespace: "Namespace",
    phase: "Phase",
    activeContainers: "Active Containers",
    totalContainers: "Total Containers",
    creationTime: "Creation Time",
    createdBy: "created By",
    action: "Action"
  };
  const initialExpandedRepoNames = pods
    .filter((repo) => !!repo.nestedComponent)
    .map((repo) => repo.name); // Default to all expanded
  const [expandedRepoNames, setExpandedRepoNames] = React.useState<string[]>(
    initialExpandedRepoNames
  );
  const setRepoExpanded = (repo: Pod, isExpanding = true) =>
    setExpandedRepoNames((prevExpanded) => {
      const otherExpandedRepoNames = prevExpanded.filter(
        (r) => r !== repo.name
      );
      return isExpanding
        ? [...otherExpandedRepoNames, repo.name]
        : otherExpandedRepoNames;
    });
  const isRepoExpanded = (repo: Pod) =>
    expandedRepoNames.includes(repo.name);

  const defaultActions = (repo: Pod): IAction[] => [
    {
      title: "Show Logs",
      onClick: () => fetchDSLQuery(repo.uuid)
    },
    {
      isSeparator: true
    },
    {
      title: "Third action",
      onClick: () => console.log(`clicked on Third action, on row ${repo.name}`)
    }
  ];
  const tableToolbar = (
    <Toolbar usePageInsets id="compact-toolbar">
      <ToolbarContent>
        
        <ToolbarItem variant="pagination">
          {renderPagination("top", true)}
        </ToolbarItem>
      </ToolbarContent>
    </Toolbar>
  );

  const generateTableCells = (repo) => (
    Object.keys(columnNames).map((key, index) => {
      if (key === "action") {
        return (
          <Th dataLabel={columnNames.action}>
            <ActionsColumn items={defaultActions(repo)} />
          </Th>
        );
      } else {
        return (
          <Th
          //  modifier="breakWord"
            modifier="wrap"
            dataLabel={columnNames[key]}
          >
            {repo[key].toString()}
          </Th>
        );
      }
    }))

  const renderTableRows = () => (
     
    paginatedRows.map((repo, rowIndex) => {
        repo.nestedComponent = <PodTabs name={repo.name} namespace={repo.namespace} uuid={repo.uuid}/>
        return (
        <Tbody key={repo.name} isExpanded={isRepoExpanded(repo)}>
          <Tr>
          <Td
              expand={
                repo.nestedComponent
                  ? {
                      rowIndex,
                      isExpanded: isRepoExpanded(repo),
                      onToggle: () => setRepoExpanded(repo, !isRepoExpanded(repo)),
                      expandId: 'composable-nested-table-expandable-example'
                    }
                  : undefined
              }
            />
          {generateTableCells(repo)}
        </Tr>
        {repo.nestedComponent ? (
            <Tr isExpanded={isRepoExpanded(repo)}>
              <Td
                noPadding={repo.noPadding}
                dataLabel={`${columnNames.name} expended`}
                colSpan={Object.keys(columnNames).length + 1}
              >
                <ExpandableRowContent>{isRepoExpanded(repo) ? repo.nestedComponent : null }</ExpandableRowContent>
              </Td>
            </Tr>
          ) : null}
      </Tbody>
        )}
    )
    
    )

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






const tableComposable = (
    <TableComposable variant="compact" aria-label="Simple table">
      <Thead>
        <Tr>
          <Td />
          {Object.keys(columnNames).map((key, index) => {
            return <Th modifier="wrap">{columnNames[key]}</Th>;
          })}
        </Tr>
      </Thead>
      { loadingData ? (loadingElem()) : (renderTableRows())}
    </TableComposable>
);

  return (
    <PageSection>
        <Card>
            {tableToolbar}
            <Divider />
            {tableComposable}
            {renderPagination("bottom", false)}
        </Card>
    </PageSection>

);
}

export { Pods };
