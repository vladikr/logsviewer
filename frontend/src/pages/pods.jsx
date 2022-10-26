import React from 'react';
import {DashboardLayout} from '../components/Layout';
import {
  QueryClient,
  QueryClientProvider
} from "@tanstack/react-query";
import {PodsTable} from './podsTable';


const queryClient = new QueryClient();
const PodsPage = () => {
  return (
    <DashboardLayout>
      <React.StrictMode>
		<QueryClientProvider client={queryClient}>
		  <PodsTable />
		</QueryClientProvider>
      </React.StrictMode>
    </DashboardLayout>
  )
}

export default PodsPage;
