import React from "react";
import { BrowserRouter, Switch, Route } from "react-router-dom";

import SettingsPage from "./settings";
import WorkloadsPage from "./workloads";
import NodesPage from "./nodes";
import ImportPage from "./import";
import PodsPage from "./pods";
import PodsExpPage from "./podsExp";
import VMIsPage from "./vmis";
import VMIMigrationsPage from "./vmims";
import VMsPage from "./vms";
import ImportLogsPage from "./importlogs";
import ImportObservedbPage from "./importdb";
import HomePage from "./home";

const Routes = () => {
  return (
    <BrowserRouter>
      <Switch>
        <Route path="/import/logs">
          <ImportLogsPage />
        </Route>
        <Route path="/import/observedb">
          <ImportObservedbPage />
        </Route>
        <Route path="/workloads/vms">
          <VMsPage />
        </Route>
        <Route path="/workloads/vmis">
          <VMIsPage />
        </Route>
        <Route path="/workloads/vmims">
          <VMIMigrationsPage />
        </Route>
        <Route path="/workloads/pods">
          <PodsPage />
        </Route>
        <Route path="/workloads/podsExp">
          <PodsExpPage />
        </Route>

        <Route path="/nodes">
          <NodesPage />
        </Route>
        <Route path="/settings">
          <SettingsPage />
        </Route>
        <Route path="/">
          <HomePage />
        </Route>
      </Switch>
    </BrowserRouter>
  );
};

export default Routes;
