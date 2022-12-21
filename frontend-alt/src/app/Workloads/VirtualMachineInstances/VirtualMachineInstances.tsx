import * as React from 'react';
import "@patternfly/react-core/dist/styles/base.css";
import axios from 'axios';
import { Migrations } from '@app/Workloads/Migrations/Migrations';
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
  const generateVmiData = (unproccessedData: any[]) => {
    const vmis: Vmi[] = [];
    unproccessedData.map((res) => {
      //res['cretionTime'] = new Date(res.creationTime);
      const newRes: Vmi = { ...res, creationTime: new Date(res.creationTime) };
      vmis.push(newRes);
      return vmis;
    });
    console.log(vmis);
    return vmis;
  };
  //const vmis: Vmi[] = generateVmiData();
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
      title: "Open Logs",
      onClick: () => console.log(`clicked on Some action, on row ${repo.name}`)
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
            modifier="breakWord"
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
        repo.nestedComponent = <Migrations name={repo.name} namespace={repo.namespace}/>
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
                <ExpandableRowContent>{repo.nestedComponent}</ExpandableRowContent>
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
        <Td />
        <Tr>
          {Object.keys(columnNames).map((key, index) => {
            return <Th>{columnNames[key]}</Th>;
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
