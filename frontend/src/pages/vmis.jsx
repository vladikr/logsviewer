import React from 'react';
import {DashboardLayout} from '../components/Layout';
import {
  QueryClient,
  QueryClientProvider
} from "@tanstack/react-query";
import {VmisTable} from './vmisTable';

const queryClient = new QueryClient();
const VMIsPage = () => {
  return (
    <DashboardLayout>
      <React.StrictMode>
		<QueryClientProvider client={queryClient}>
		  <VmisTable />
		</QueryClientProvider>
      </React.StrictMode>
    </DashboardLayout>
  )
}

export default VMIsPage;
