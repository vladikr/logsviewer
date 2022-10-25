import React, { useState, useEffect } from 'react';
import axios from 'axios';
import {DashboardLayout} from '../components/Layout';
import "./index.css";

//3 TanStack Libraries!!!
import {
  ColumnDef,
  ColumnSort,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  Row,
  SortingState,
  useReactTable
} from "@tanstack/react-table";
import {
  QueryClient,
  QueryClientProvider,
  useInfiniteQuery
} from "@tanstack/react-query";
import { useVirtual } from "react-virtual";
export type Pod = {
	uuid: string;
	name: string;
	namespace: string;
	phase: string;
	activeContainers: number;
	totalContainers: number;
	creationTime: Date;
};

export type PodsApiResponse = {
    data: Pod[];
    meta: {
  	    totalRowCount: number;
    }
};

const fetchSize = 25;

const queryClient = new QueryClient();
const PodsPage = () => {
  return (
    <DashboardLayout>
		<QueryClientProvider client={queryClient}>
		  <App />
		</QueryClientProvider>
    </DashboardLayout>
  )
}

export default PodsPage;


function App() {
  const [data, setData] = useState([]);
  const rerender = React.useReducer(() => ({}), {})[1];

  //we need a reference to the scrolling element for logic down below
  const tableContainerRef = React.useRef<HTMLDivElement>(null);

  //const [sorting, setSorting] = React.useState < SortingState > [];

  const columns = React.useMemo<ColumnDef<Pod>[]>(
    () => [
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
        accessorKey: "activeContainers",
        header: () => <span>Active Containers</span>,
        size: 50
      },
      {
        accessorKey: "totalContainers",
        header: () => <span>Total Containers</span>,
        size: 50
      },
      {
        accessorKey: "creationTime",
        header: "Created At",
        cell: (info) => info.getValue()
      }
    ],
    []
  );
    const fetchData = (
		start: number,
		size: number
  		//sorting: SortingState
	) => {
/*
       useEffect(() => {
           (async () => {
             //const result = await axios("https://api.tvmaze.com/search/shows?q=snow");
             const result = await axios("/pods");
             setData(result.data);
           })();
       }, []);
*/
		/*async () => {
		  const result = await axios("/pods");
		  setData(result.data);
		};*/
        return axios.get("/pods",
            {
                params: {
                    page: start,
                    per_page: size
                }
            }).then(function (resp) {
                setData(data);
                return {
                    data: data.slice(start, start + size), 
                    meta: resp.meta,
                };
            })
    }

/*
            .then((data, meta) => {
        setCurrentItems(data.slice(itemOffset, endOffset));
        setData(data);
  		const dbData = [...data]
      });
  		/*if (sorting.length) {
			const sort = sorting[0] as ColumnSort;
			const { id, desc } = sort as { id: keyof Pod; desc: boolean }
			dbData.sort((a, b) => {
				if (desc) {
					return a[id] < b[id] ? 1 : -1
				}
				return a[id] > b[id] ? 1 : -1
			})
  		}*/
/*
	  return {
		data: dbData.slice(start, start + size),
		meta: {
		  totalRowCount: dbData.length,
		},
	  }
	}
*/
  //react-query has an useInfiniteQuery hook just for this situation!
  const { fetchNextPage, isFetching, isLoading } = useInfiniteQuery<
    PodsApiResponse
  >(
    ["table-data"], //adding sorting state as key causes table to reset and fetch from new beginning upon sort
	async ({ pageParam = 0 }) => {
      const start = pageParam * fetchSize;
      const fetchedData = fetchData(start, fetchSize); //, sorting);
      return fetchedData;
    },
    {
      getNextPageParam: (_lastGroup, groups) => groups.length,
      keepPreviousData: true,
      refetchOnWindowFocus: false
    }
  );


//    useEffect(() => {
//    (async () => {
//      const result = await axios("https://api.tvmaze.com/search/shows?q=snow");
//      setData(result.data);
//    })();
 // }, []);
  //we must flatten the array of arrays from the useInfiniteQuery hook
  const flatData = React.useMemo(
    () => data?.pages?.flatMap((page) => page.data) ?? [],
    [data]
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
    //state: {
    //  sorting
    //},
    //onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    debugTable: true
  });

  const { rows } = table.getRowModel();

  //Virtualizing is optional, but might be necessary if we are going to potentially have hundreds or thousands of rows
  const rowVirtualizer = useVirtual({
    parentRef: tableContainerRef,
    size: rows.length,
    overscan: 10
  });
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
        onScroll={(e) => fetchMoreOnBottomReached(e.target)}
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
                          }[header.column.getIsSorted()] ?? null}
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
              const row = rows[virtualRow.index]; // as Row<Pod>;
              return (
                <tr key={row.id}>
                  {row.getVisibleCells().map((cell) => {
                    return (
                      <td key={cell.id}>
                        {flexRender(
                          cell.column.columnDef.cell,
                          cell.getContext()
                        )}
                      </td>
                    );
                  })}
                </tr>
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

