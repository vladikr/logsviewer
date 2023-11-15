import * as React from 'react';
import "@patternfly/react-core/dist/styles/base.css";
import axios from 'axios';
import { VirtualMachineTabs } from '@app/Workloads/VirtualMachines/VirtualMachineTabs';
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
  Card,
  Pagination,
  PageSection,
  Toolbar,
  ToolbarContent,
  ToolbarGroup,
  ToolbarItem,
  Bullseye, EmptyState, EmptyStateIcon, Spinner, Title,
} from "@patternfly/react-core";
import { apiBaseUrl } from "@app/config";
import * as queryString from "querystring";
import {useLocation} from "react-router-dom";

const VirtualMachines: React.FunctionComponent = () => {
    type Vm = {
        uuid: string;
        name: string;
        namespace: string;
        running: string;
        created: string;
        ready: string;
        status: string;
        nestedComponent?: React.ReactNode;
        link?: React.ReactNode;
        noPadding?: boolean;
    };

    const { search } = useLocation()
    const queryStringValues = queryString.parse(search.slice(1))

	const [loadingData, setLoadingData] = React.useState(true);
  	const [data, setData] = React.useState<any[]>([]);

  	React.useEffect(() => {
      let url = apiBaseUrl + "/vms";

      if (queryStringValues.status !== undefined) {
        url = url + "?status=" + queryStringValues.status
      }

    	async function getData() {
      	await axios
        	.get(url)
        	.then((response) => {
          	// check if the data is populated
            const processedData = generateVmData(response.data.data);
          	setData(processedData);
            const localVms = processedData
            setPaginatedRows(localVms.slice(0, 10));
          	console.log(paginatedRows);
          	setLoadingData(false);
        	});
    	}
        if (loadingData) {
        // if the result is not ready so you make the axios call
            getData();
        }
    }, []);

    const generateVmData = (unproccessedData: any[]) => {
        const vms: Vm[] = [];
        unproccessedData.map((res) => {
        const newRes: Vm = { ...res };
            vms.push(newRes);
            return vms;
        });
        return vms;
    };

  const vms: Vm[] = data;
  const [page, setPage] = React.useState(1);
  const [perPage, setPerPage] = React.useState(10);
  const [paginatedRows, setPaginatedRows] = React.useState(vms.slice(0, 10));
  const handleSetPage = (_evt, newPage, perPage, startIdx, endIdx) => {
    setPaginatedRows(vms.slice(startIdx, endIdx));
    setPage(newPage);
  };
  const handlePerPageSelect = (_evt, newPerPage, newPage, startIdx, endIdx) => {
    setPaginatedRows(vms.slice(startIdx, endIdx));
    setPage(newPage);
    setPerPage(newPerPage);
  };

    const fetchDSLQuery = async (
		vmUUID: string,
		nodeName: string,
        apiPath: string,
	) => {
        const retq = await axios.get(apiBaseUrl + "/" + apiPath,
            {
                params: {
                    vmUUID: vmUUID,
                    nodeName: nodeName
                }
            }).then(function (resp) {
                console.log("await2: ", resp.data.dslQuery)
                const hostname = window.location.hostname
                const hostnameParts = hostname.split('.');
                const ingress = hostnameParts.slice(1).join('.');
                const appNameParts = hostnameParts.slice(0, 1)[0].split('-');
                let prefix = ""
                let suffix = ""
                if (appNameParts.length > 1) {
                    prefix = appNameParts[0] + "-";
                    suffix = "-" + appNameParts.slice(2);
                }
                const kibanaHostname = prefix + "kibana" + suffix + "." + ingress;

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
      itemCount={vms.length}
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
    running: "Running",
    created: "Created",
    ready: "Ready",
    status: "Status",
    action: "Action"
  };

  const initialExpandedRepoNames = vms
    .filter((repo) => !!repo.nestedComponent)
    .map((repo) => repo.name); // Default to all expanded
  const [expandedRepoNames, setExpandedRepoNames] = React.useState<string[]>(
    initialExpandedRepoNames
  );
  const setRepoExpanded = (repo: Vm, isExpanding = true) =>
    setExpandedRepoNames((prevExpanded) => {
      const otherExpandedRepoNames = prevExpanded.filter(
        (r) => r !== repo.name
      );
      return isExpanding
        ? [...otherExpandedRepoNames, repo.name]
        : otherExpandedRepoNames;
    });
  const isRepoExpanded = (repo: Vm) =>
    expandedRepoNames.includes(repo.name);

  const defaultActions = (repo: Vm): IAction[] => [
    {
      title: "Show Logs",
      onClick: () => console.log("VM logs are not implemented yet")
    },
    {
      isSeparator: true
    },
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
            modifier="wrap"
            dataLabel={columnNames[key]}
          >
            {repo[key] !== undefined ? repo[key].toString() : ""}
          </Th>
        );
      }
    }))

  const renderTableRows = () => {
    const newDataRows = paginatedRows
    return (
    
    newDataRows.map((repo, rowIndex) => { 
        repo.nestedComponent = <VirtualMachineTabs name={repo.name} namespace={repo.namespace} uuid={repo.uuid} />
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


    return (
    <PageSection>
    <Card>
      {tableToolbar}
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
    {renderPagination("bottom", false)}
    </Card>
  </PageSection>
);
}
  
export { VirtualMachines };
