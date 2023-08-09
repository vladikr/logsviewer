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
import { PVCTabs } from '@app/Storage/PVC/PVCTabs';
import { apiBaseUrl } from "@app/config";
import {useLocation} from "react-router-dom";
import * as queryString from "querystring";

const PersistentVolumeClaims: React.FunctionComponent = () => {
  const { search } = useLocation()
  const queryStringValues = queryString.parse(search.slice(1))

	const [loadingData, setLoadingData] = React.useState(true);
  	const [data, setData] = React.useState<any[]>([]);

  	React.useEffect(() => {
      let url = apiBaseUrl + "/getPVCs"

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
            const localPvcs = processedData
            setPaginatedRows(localPvcs.slice(0, 10));
          	// you tell it that you had the result
          	setLoadingData(false);
        	});
    	}
    if (loadingData) {
      // if the result is not ready so you make the axios call
      getData();
    }
  }, []);

  type PVC = {
    name: string;
    namespace: string;
    uuid: string;
    accessModes: string;
    storageClassName: string;
    volumeName: string;
    volumeMode: string;
    phase: string;
    capacity: string;
    creationTime: Date;
    nestedComponent?: React.ReactNode;
    link?: React.ReactNode;
    noPadding?: boolean;
  };
  const generatePodData = (unproccessedData: any[]) => {
    const pvcs: PVC[] = [];
    unproccessedData.map((res) => {
      const newRes: PVC = { ...res, creationTime: new Date(res.creationTime) };
      pvcs.push(newRes);
      return pvcs;
    });
    console.log(pvcs);
    return pvcs;
  };

  const pvcs: PVC[] = data;

  const [page, setPage] = React.useState(1);
  const [perPage, setPerPage] = React.useState(10);
  const [paginatedRows, setPaginatedRows] = React.useState(pvcs.slice(0, 10));
  const handleSetPage = (_evt, newPage, perPage, startIdx, endIdx) => {
    setPaginatedRows(pvcs.slice(startIdx, endIdx));
    setPage(newPage);
  };
  const handlePerPageSelect = (_evt, newPerPage, newPage, startIdx, endIdx) => {
    setPaginatedRows(pvcs.slice(startIdx, endIdx));
    setPage(newPage);
    setPerPage(newPerPage);
  };

  const renderPagination = (variant, isCompact) => (
    <Pagination
      isCompact={isCompact}
      itemCount={pvcs.length}
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
    storageClassName: "StorageClass",
    volumeName: "Volume",
    volumeMode: "VolumeMode",
    phase: "Phase",
    capacity: "Capacity",
    creationTime: "CreationTime",
    action: "Action"
  };
  const initialExpandedRepoNames = pvcs
    .filter((repo) => !!repo.nestedComponent)
    .map((repo) => repo.name); // Default to all expanded
  const [expandedRepoNames, setExpandedRepoNames] = React.useState<string[]>(
    initialExpandedRepoNames
  );
  const setRepoExpanded = (repo: PVC, isExpanding = true) =>
    setExpandedRepoNames((prevExpanded) => {
      const otherExpandedRepoNames = prevExpanded.filter(
        (r) => r !== repo.name
      );
      return isExpanding
        ? [...otherExpandedRepoNames, repo.name]
        : otherExpandedRepoNames;
    });
  const isRepoExpanded = (repo: PVC) =>
    expandedRepoNames.includes(repo.name);

  const defaultActions = (repo: PVC): IAction[] => [
    {
      title: "PVC",
      onClick: () => console.log(`Not Implemented yet`) 
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
        repo.nestedComponent = <PVCTabs uuid={repo.uuid} />
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

export { PersistentVolumeClaims };
