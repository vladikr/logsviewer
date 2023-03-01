import * as React from 'react';
import "@patternfly/react-core/dist/styles/base.css";
import axios from 'axios';
import { VirtualMachineInstancesTabs } from '@app/Workloads/VirtualMachineInstances/VirtualMachineInstancesTabs';
//import { Migrations } from '@app/Workloads/Migrations/Migrations';
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

const VirtualMachineInstances: React.FunctionComponent = () => {
    type Vmi = {
        uuid: string;
        name: string;
        namespace: string;
        phase: string;
        reason: string;
        nodeName: string;
        creationTime: Date;
        nestedComponent?: React.ReactNode;
        link?: React.ReactNode;
        noPadding?: boolean;
    };

	const [loadingData, setLoadingData] = React.useState(true);
  	const [data, setData] = React.useState<any[]>([]);

  	React.useEffect(() => {
    	async function getData() {
      	await axios
        	.get("/vmis")
        	.then((response) => {
          	// check if the data is populated
          	console.log(response.data);
            const processedData = generateVmiData(response.data.data);
          	setData(processedData);
            const localVmis = processedData
            setPaginatedRows(localVmis.slice(0, 10));
          	// you tell it that you had the result
          	setLoadingData(false);
        	});
    	}
        if (loadingData) {
        // if the result is not ready so you make the axios call
            getData();
        }
    }, []);

    const generateVmiData = (unproccessedData: any[]) => {
        const vmis: Vmi[] = [];
        unproccessedData.map((res) => {
        const newRes: Vmi = { ...res, creationTime: new Date(res.creationTime) };
            vmis.push(newRes);
            return vmis;
        });
        return vmis;
    };

  const vmis: Vmi[] = data;
  const [page, setPage] = React.useState(1);
  const [perPage, setPerPage] = React.useState(10);
  const [paginatedRows, setPaginatedRows] = React.useState(vmis.slice(0, 10));
  const handleSetPage = (_evt, newPage, perPage, startIdx, endIdx) => {
    setPaginatedRows(vmis.slice(startIdx, endIdx));
    setPage(newPage);
  };
  const handlePerPageSelect = (_evt, newPerPage, newPage, startIdx, endIdx) => {
    setPaginatedRows(vmis.slice(startIdx, endIdx));
    setPage(newPage);
    setPerPage(newPerPage);
  };

    const fetchDSLQuery = async (
		vmiUUID: string,
		nodeName: string
	) => {
        const retq = await axios.get("/getVMIQueryParams",
            {
                params: {
                    vmiUUID: vmiUUID,
                    nodeName: nodeName
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
    //const openInNewTab = ({ row }: { row: Row<Vmi> }) => {
    //    fetchDSLQuery(row.original.uuid, row.original.nodeName);
    //}

  const renderPagination = (variant, isCompact) => (
    <Pagination
      isCompact={isCompact}
      itemCount={vmis.length}
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
    reason: "Reason",
    creationTime: "Creation Time",
    nodeName: "Node",
    action: "Action"
  };
  // In this example, expanded rows are tracked by the repo names from each row. This could be any unique identifier.
  // This is to prevent state from being based on row order index in case we later add sorting.
  // Note that this behavior is very similar to selection state.
  const initialExpandedRepoNames = vmis
    .filter((repo) => !!repo.nestedComponent)
    .map((repo) => repo.name); // Default to all expanded
  const [expandedRepoNames, setExpandedRepoNames] = React.useState<string[]>(
    initialExpandedRepoNames
  );
  const setRepoExpanded = (repo: Vmi, isExpanding = true) =>
    setExpandedRepoNames((prevExpanded) => {
      const otherExpandedRepoNames = prevExpanded.filter(
        (r) => r !== repo.name
      );
      return isExpanding
        ? [...otherExpandedRepoNames, repo.name]
        : otherExpandedRepoNames;
    });
  const isRepoExpanded = (repo: Vmi) =>
    expandedRepoNames.includes(repo.name);

  const defaultActions = (repo: Vmi): IAction[] => [
    {
      title: "Show Logs",
      onClick: () => fetchDSLQuery(repo.uuid, repo.nodeName)
    },
    {
      isSeparator: true
    },
    {
      title: "Show logs since...",
      onClick: () => console.log(`Not Implemented yet for ${repo.name}`)
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
            modifier="wrap"
            dataLabel={columnNames[key]}
          >
            {repo[key].toString()}
          </Th>
        );
      }
    }))

  const renderTableRows = () => {
    const newDataRows = paginatedRows
    return (
    
    newDataRows.map((repo, rowIndex) => { 
        repo.nestedComponent = <VirtualMachineInstancesTabs name={repo.name} namespace={repo.namespace} uuid={repo.uuid} />
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
export { VirtualMachineInstances };
