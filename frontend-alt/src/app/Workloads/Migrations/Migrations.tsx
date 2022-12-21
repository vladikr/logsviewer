import * as React from 'react';
import "@patternfly/react-core/dist/styles/base.css";
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
  Card,
  Pagination,
  PageSection,
  Toolbar,
  ToolbarContent,
  ToolbarGroup,
  ToolbarItem,
  Bullseye, EmptyState, EmptyStateIcon, Spinner, Title,
} from "@patternfly/react-core";

interface VmiMigrationsTableProps {
    namespace?: string,
    name?: string
}

const Migrations: React.FunctionComponent<VmiMigrationsTableProps> = ({name, namespace}: VmiMigrationsTableProps) => { 
	const [loadingData, setLoadingData] = React.useState(true);
  	const [data, setData] = React.useState<any[]>([]);

  	React.useEffect(() => {
    	async function getData() {
      	await axios
        	.get("/vmims",
            {
                params: {
                    name: {name},
                    namespace: {namespace}
                }
            })
        	.then((response) => {
          	// check if the data is populated
          	console.log(response.data);
            const processedData = generateVmimData(response.data.data);
          	setData(processedData);
            const localVmims = processedData
            setPaginatedRows(localVmims.slice(0, 10));
          	console.log('paginatedRows: ', paginatedRows);
          	// you tell it that you had the result
          	setLoadingData(false);
        	});
    	}
    if (loadingData) {
      // if the result is not ready so you make the axios call
      getData();
    }
  }, []);

  type VmiMigration = {
    uuid: string;
    name: string;
    namespace: string;
    phase: string;
    vmiName: string;
    targetPod: string;
    creationTime: Date;
    endTimestamp: Date;
    sourceNode: string;
    targetNode: string;
    completed: boolean;
    failed: boolean;
    nestedComponent?: React.ReactNode;
    link?: React.ReactNode;
    noPadding?: boolean;
  };

  const generateVmimData = (unproccessedData: any[]) => {
    const vmims: VmiMigration[] = [];
    unproccessedData.map((res) => {
      //res['cretionTime'] = new Date(res.creationTime);
      const newRes: VmiMigration = { ...res, creationTime: new Date(res.creationTime) };
      vmims.push(newRes);
      return vmims;
    });
    console.log(vmims);
    return vmims;
  };
  //const vmims: VmiMigration[] = generateVmimData();
  const vmims: VmiMigration[] = data;

  const [page, setPage] = React.useState(1);
  const [perPage, setPerPage] = React.useState(10);
  const [paginatedRows, setPaginatedRows] = React.useState(vmims.slice(0, 10));
  const handleSetPage = (_evt, newPage, perPage, startIdx, endIdx) => {
    setPaginatedRows(vmims.slice(startIdx, endIdx));
    setPage(newPage);
  };
  const handlePerPageSelect = (_evt, newPerPage, newPage, startIdx, endIdx) => {
    setPaginatedRows(vmims.slice(startIdx, endIdx));
    setPage(newPage);
    setPerPage(newPerPage);
  };

  const renderPagination = (variant, isCompact) => (
    <Pagination
      isCompact={isCompact}
      itemCount={vmims.length}
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
    vmiName: "VMI Name",
    targetPod: "Target Pod",
    creationTime: "Creation Time",
    endTimestamp: "Ended At",
    sourceNode: "Source Node",
    targetNode: "Target Node",
    completed: "Completed",
    failed: "Failed",
    action: "Action"
  };
  // In this example, expanded rows are tracked by the repo names from each row. This could be any unique identifier.
  // This is to prevent state from being based on row order index in case we later add sorting.
  // Note that this behavior is very similar to selection state.
  const initialExpandedRepoNames = vmims
    .filter((repo) => !!repo.nestedComponent)
    .map((repo) => repo.name); // Default to all expanded
  const [expandedRepoNames, setExpandedRepoNames] = React.useState<string[]>(
    initialExpandedRepoNames
  );
  const setRepoExpanded = (repo: VmiMigration, isExpanding = true) =>
    setExpandedRepoNames((prevExpanded) => {
      const otherExpandedRepoNames = prevExpanded.filter(
        (r) => r !== repo.name
      );
      return isExpanding
        ? [...otherExpandedRepoNames, repo.name]
        : otherExpandedRepoNames;
    });
  const isRepoExpanded = (repo: VmiMigration) =>
    expandedRepoNames.includes(repo.name);

  const defaultActions = (repo: VmiMigration): IAction[] => [
    {
      title: "Some action",
      onClick: () => console.log(`clicked on Some action, on row ${repo.name}`)
    },
    {
      title: <a href="https://www.patternfly.org">Link action</a>
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
    const newDataRows = paginatedRows;
    console.log('newDataRows: ', newDataRows);
    return (
    newDataRows.map((repo, rowIndex) => (
        <Tbody key={repo.name}>
          <Tr>
          {generateTableCells(repo)}
        </Tr>
      </Tbody>
        )
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
);}

export { Migrations };
