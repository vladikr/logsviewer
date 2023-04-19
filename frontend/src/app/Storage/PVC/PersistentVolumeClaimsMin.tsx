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

interface PVCTableProps {
    namespace?: string,
    name?: string,
    uuid?: string,
    object: string
}

const PVCsTableMinimal: React.FunctionComponent<PVCTableProps> = ({name, namespace, uuid, object}: PVCTableProps) => {

	const [loadingData, setLoadingData] = React.useState(true);
  	const [data, setData] = React.useState<any[]>([]);

    const renderAPIGet = (object: string): string => {
      switch(object) {
        case 'pod':
          return '/getPodPVCs';
        case 'vmi':
          return '/getVMIPVCs';
        default:
          return '';
      }
      return "";
    }


  	React.useEffect(() => {
        let apiVerb = renderAPIGet(object)
    	async function getData() {
      	await axios
        	.get(apiVerb!,
            {
                params: {
                    uuid: uuid
                }
            })
        	.then((response) => {
          	// check if the data is populated
          	console.log(response.data);
            const processedData = generatePodData(response.data.data);
          	setData(processedData);
            const localPvcs = processedData
            setPaginatedRows(localPvcs.slice(0, 10));
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

export { PVCsTableMinimal };
