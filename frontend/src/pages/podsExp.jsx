import React from 'react';
import {DashboardLayout} from '../components/Layout';
import {
  QueryClient,
  QueryClientProvider
} from "@tanstack/react-query";
import {PodsExpTable} from './podsExpTable';


const queryClient = new QueryClient();
const PodsExpPage = () => {
  return (
    <DashboardLayout>
      <React.StrictMode>
		<QueryClientProvider client={queryClient}>
		  <PodsExpTable />
		</QueryClientProvider>
      </React.StrictMode>
    </DashboardLayout>
  )
}

export default PodsExpPage;
