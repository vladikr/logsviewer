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

const Pods: React.FunctionComponent = () => {

	const [loadingData, setLoadingData] = React.useState(true);
  	const [data, setData] = React.useState<any[]>([]);

  	React.useEffect(() => {
    	async function getData() {
      	await axios
        	.get("/pods")
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

  /*const formatPodData = () => {
    const kuki: Repository[] = [
      {
        name: "Node 1",
        branches: "10",
        prs: "2",
        nestedComponent: <NestedReposTable />,
        link: <a>Link 1</a>
      },
      { name: "Node 2", branches: "3", prs: "4", link: <a>Link 2</a> },
      {
        name: "Node 3",
        branches: "11",
        prs: "7",
        nestedComponent: (
          <p>
            Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do
            eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim
            ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut
            aliquip ex ea commodo consequat. Duis aute irure dolor in
            reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla
            pariatur. Excepteur sint occaecat cupidatat non proident, sunt in
            culpa qui officia deserunt mollit anim id est laborum.
          </p>
        ),
        link: <a>Link 3</a>
      },
      {
        name: "Node 4",
        branches: "11",
        prs: "7",
        nestedComponent: "Expandable row content has no padding.",
        link: <a>Link 4</a>,
        noPadding: true
      }
    ];
    console.log(kuki);

    return kuki;
  };*/

  // In real usage, this data would come from some external source like an API via props.
  //const repositories: Repository[] = formatPodData()
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
  // In this example, expanded rows are tracked by the repo names from each row. This could be any unique identifier.
  // This is to prevent state from being based on row order index in case we later add sorting.
  // Note that this behavior is very similar to selection state.
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

  const renderTableRows = () => (
     
    paginatedRows.map((repo, rowIndex) => (
        <Tbody key={repo.name}>
          <Tr>
          {generateTableCells(repo)}
        </Tr>
      </Tbody>
        )
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
);
}

export { Pods };
