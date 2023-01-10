import React from 'react';
import {DashboardLayout} from '../components/Layout';
import {
  QueryClient,
  QueryClientProvider
} from "@tanstack/react-query";
import {VmiMigrationsTable} from './vmimsTable';

const queryClient = new QueryClient();
const VMIMigrationsPage = () => {
  return (
    <DashboardLayout>
      <React.StrictMode>
		<QueryClientProvider client={queryClient}>
		  <VmiMigrationsTable />
		</QueryClientProvider>
      </React.StrictMode>
    </DashboardLayout>
  )
}

export default VMIMigrationsPage;
