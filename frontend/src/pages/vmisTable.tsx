import React, { useState, Fragment } from 'react';
import axios from 'axios';
//import { Pod, PodsApiResponse } from './podObject';
import "./index.css";
import {VmiMigrationsTable} from './vmimsTable';
import Icon from "awesome-react-icons";

//3 TanStack Libraries!!!
import {
  ColumnDef,
  Row,
  flexRender,
  getExpandedRowModel,
  getCoreRowModel,
  getSortedRowModel,
  useReactTable
} from "@tanstack/react-table";
import {
  QueryClient,
  QueryClientProvider,
  useInfiniteQuery
} from "@tanstack/react-query";
import { useVirtual } from "react-virtual";
export type Vmi = {
    uuid: string;
    name: string;
    namespace: string;
    phase: string;
    reason: string;
    nodeName: string;
    creationTime: Date;
};

type VmisApiResponse = {
    data: Vmi[];
    meta: {
        totalRowCount: number;
    }
};


const fetchSize = 25;

export const VmisTable = () => {
    const rerender = React.useReducer(() => ({}), {})[1];
    //we need a reference to the scrolling element for logic down below
    const tableContainerRef = React.useRef<HTMLDivElement>(null);
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
                const appNameParts = hostnameParts.slice(0, 1).split('-');
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
    const openInNewTab = ({ row }: { row: Row<Vmi> }) => {
        fetchDSLQuery(row.original.uuid, row.original.nodeName);
    }
    const columns = React.useMemo<ColumnDef<Vmi>[]>(
        () => [
          {
            id: 'reporter',
            header: () => null,
            size: 20,
            cell: ({ row }) => {
                return (
                    <button
                        {...{
                            onClick: () => openInNewTab({row}),
                            style: { cursor: 'pointer' },
                        }}
                    >   <Icon name="external-link"/>
                    </button>
                )
            },
          },
          {
            id: 'expander',
            header: () => null,
            size: 20,
            cell: ({ row }) => {
                return row.getCanExpand() ? (
                    <button
                        {...{
                            onClick: row.getToggleExpandedHandler(),
                            style: { cursor: 'pointer' },
                        }}
                    >
                        {row.getIsExpanded() ? <Icon name="minus"/> : <Icon name="plus"/>}
                    </button>
                ) : (
                    '🔵'
                )
            },
          },
          {
            accessorKey: "uuid",
            cell: (info) => info.getValue(),
            header: () => <span>UUID</span>
          },
          {
            accessorKey: "name",
            cell: (info) => info.getValue(),
            header: () => <span>Name</span>
          },
          {
            accessorKey: "namespace",
            cell: (info) => info.getValue(),
            header: () => <span>Namespace</span>
          },
          {
            accessorKey: "phase",
            cell: (info) => info.getValue(),
            header: () => <span>Phase</span>
          },
          {
            accessorKey: "reason",
            cell: (info) => info.getValue(),
            header: () => <span>Reason</span>
          },
          {
            accessorKey: "creationTime",
            header: "Created At",
            cell: (info) => info.getValue()
          },
          {
            accessorKey: "nodeName",
            cell: (info) => info.getValue(),
            header: () => <span>NodeName</span>
          }
        ],
        []
    );

    const fetchData = (
		start: number,
		size: number
	) => {
        return axios.get("/vmis",
            {
                params: {
                    page: start,
                    per_page: size
                }
            }).then(function (resp) {
                return {
                    data: resp.data.data, 
                    meta: resp.data.meta,
                };
            })
    }

  //react-query has an useInfiniteQuery hook just for this situation!
  const { data, fetchNextPage, isFetching, isLoading } = useInfiniteQuery<
    VmisApiResponse
  >(
    ["table-data"], //adding sorting state as key causes table to reset and fetch from new beginning upon sort
	async ({ pageParam = 0 }) => {
      const start = pageParam;
      const fetchedData = fetchData(start, fetchSize);
      return fetchedData;
    },
    {
      getNextPageParam: (_lastGroup, groups) => groups.length,
      keepPreviousData: true,
      refetchOnWindowFocus: false
    }
  );

  //we must flatten the array of arrays from the useInfiniteQuery hook
  const flatData = React.useMemo(
    () => {
        const mData = data?.pages?.flatMap((page) => page.data) ?? []
        if (data === undefined || mData.length === 0) {
            return mData;
        } 
        return Object.values(
            mData.reduce<Record<string, Vmi>>((c, v) => {
            if (v.hasOwnProperty('uuid')) {
                c[v.uuid] = v;
            }
            return c;
          }, {}));
        
    }, [data]
  );
  const totalDBRowCount = data?.pages?.[0]?.meta?.totalRowCount ?? 0;
  const totalFetched = flatData.length;

  //called on scroll and possibly on mount to fetch more data as the user scrolls and reaches bottom of table
  const fetchMoreOnBottomReached = React.useCallback(
    (containerRefElement?: HTMLDivElement | null) => {
      if (containerRefElement) {
        const { scrollHeight, scrollTop, clientHeight } = containerRefElement;
        //once the user has scrolled within 300px of the bottom of the table, fetch more data if there is any
        if (
          scrollHeight - scrollTop - clientHeight < 300 &&
          !isFetching &&
          totalFetched < totalDBRowCount
        ) {
          fetchNextPage();
        }
      }
    },
    [fetchNextPage, isFetching, totalFetched, totalDBRowCount]
  );

  //a check on mount and after a fetch to see if the table is already scrolled to the bottom and immediately needs to fetch more data
  React.useEffect(() => {
    fetchMoreOnBottomReached(tableContainerRef.current);
  }, [fetchMoreOnBottomReached]);

  const table = useReactTable({
    data: flatData,
    columns,
    getRowCanExpand: () => true,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getExpandedRowModel: getExpandedRowModel(),
    debugTable: true
  });

  const { rows } = table.getRowModel();

  //Virtualizing is optional, but might be necessary if we are going to potentially have hundreds or thousands of rows
  const rowVirtualizer = useVirtual({
    parentRef: tableContainerRef,
    size: rows.length,
    overscan: 10
  });
  
  const renderSubComponent = ({ row }: { row: Row<Vmi> }) => {
    const queryClient = new QueryClient();
  return (
    
      <React.StrictMode>
		<QueryClientProvider client={queryClient}>
            <VmiMigrationsTable name={row.original.name} namespace={row.original.namespace}/>
		</QueryClientProvider>
      </React.StrictMode>
    /*<pre style={{ fontSize: '10px' }}>
      <code>{JSON.stringify(row.original, null, 2)}</code>
      
    </pre>*/
  )
}
  const { virtualItems: virtualRows, totalSize } = rowVirtualizer;
  const paddingTop = virtualRows.length > 0 ? virtualRows?.[0]?.start || 0 : 0;
  const paddingBottom =
    virtualRows.length > 0
      ? totalSize - (virtualRows?.[virtualRows.length - 1]?.end || 0)
      : 0;

  if (isLoading) {
    return <>Loading...</>;
  }
//as HTMLDivElement
  return (
    <div className="p-2">
      <div className="h-2" />
      <div
        className="container"
        onScroll={(e) => fetchMoreOnBottomReached(e.target as HTMLDivElement)}
        ref={tableContainerRef}
      >
        <table>
          <thead>
            {table.getHeaderGroups().map((headerGroup) => (
              <tr key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <th
                      key={header.id}
                      colSpan={header.colSpan}
                      style={{ width: header.getSize() }}
                    >
                      {header.isPlaceholder ? null : (
                        <div
                          {...{
                            className: header.column.getCanSort()
                              ? "cursor-pointer select-none"
                              : "",
                            onClick: header.column.getToggleSortingHandler()
                          }}
                        >
                          {flexRender(
                            header.column.columnDef.header,
                            header.getContext()
                          )}
                          {{
                            asc: " 🔼",
                            desc: " 🔽"
                          }[header.column.getIsSorted() as string] ?? null}
                        </div>
                      )}
                    </th>
                  );
                })}
              </tr>
            ))}
          </thead>
          <tbody>
            {paddingTop > 0 && (
              <tr>
                <td style={{ height: `${paddingTop}px` }} />
              </tr>
            )}
            {virtualRows.map((virtualRow) => {
              const row = rows[virtualRow.index] as Row<Vmi>;
              return (
                <Fragment key={row.id}>
                    <tr>
                      {/* first row is a normal row */}
                      {row.getVisibleCells().map(cell => {
                        return (
                          <td key={cell.id}>
                            {flexRender(
                              cell.column.columnDef.cell,
                              cell.getContext()
                            )}
                          </td>
                        )
                      })}
                    </tr>
                    {row.getIsExpanded() && (
                      <tr>
                        {/* 2nd row is a custom 1 cell row */}
                        <td colSpan={row.getVisibleCells().length}>
                          {renderSubComponent({ row })}
                        </td>
                      </tr>
                    )}
                  </Fragment>
              );
            })}
            {paddingBottom > 0 && (
              <tr>
                <td style={{ height: `${paddingBottom}px` }} />
              </tr>
            )}
          </tbody>
        </table>
      </div>
      <div>
        Fetched {flatData.length} of {totalDBRowCount} Rows.
      </div>
      <div>
        <button onClick={() => rerender()}>Force Rerender</button>
      </div>
    </div>
  );
}
