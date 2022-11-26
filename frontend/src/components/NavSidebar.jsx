/* eslint-disable react/display-name, jsx-a11y/click-events-have-key-events */
import { Navigation } from "react-minimal-side-navigation";
import { useHistory, useLocation } from "react-router-dom";
import Icon from "awesome-react-icons";
import React, { useState } from "react";

import "react-minimal-side-navigation/lib/ReactMinimalSideNavigation.css";

export const NavSidebar = () => {
  const history = useHistory();
  const location = useLocation();
  const [isSidebarOpen, setIsSidebarOpen] = useState(true);

  return (
    <React.Fragment>
      {/* Sidebar Overlay */}
      <div
        onClick={() => setIsSidebarOpen(false)}
        className={`fixed inset-0 z-20 block transition-opacity bg-black opacity-50 lg:hidden ${
          isSidebarOpen ? "block" : "hidden"
        }`}
      />

      <div className="absolute right-0">
        <a href="https://github.com/abhijithvijayan/react-minimal-side-navigation">
        </a>
      </div>

      <div>
        <button
          className="btn-menu"
          onClick={(): void => setIsSidebarOpen(true)}
          type="button"
        >
          <Icon name="burger" className="w-6 h-6" />
        </button>
      </div>

      {/* Sidebar */}
      <div
        className={`fixed inset-y-0 left-0 z-30 w-64 overflow-y-auto transition duration-300 ease-out transform translate-x-0 bg-white border-r-2 lg:translate-x-0 lg:static lg:inset-0 ${
          isSidebarOpen ? "ease-out translate-x-0" : "ease-in -translate-x-full"
        }`}
      >
        

        {/* https://github.com/abhijithvijayan/react-minimal-side-navigation */}
        <Navigation
          activeItemId={location.pathname}
          onSelect={({ itemId }) => {
            history.push(itemId);
          }}
          items={[
            {
              title: "Home",
              itemId: "/home",
              // Optional
              elemBefore: () => <Icon name="circle" />
            },
            {
              title: "Import",
              itemId: "/import",
              elemBefore: () => <Icon name="briefcase" />,
              subNav: [
                {
                  title: "Logs",
                  itemId: "/import/logs",
                  // Optional
                  elemBefore: () => <Icon name="plus" />
                },
                {
                  title: "ObserveDB",
                  itemId: "/import/observedb",
                  elemBefore: () => <Icon name="plus" />
                }
              ]
            },
            {
              title: "Workloads",
              itemId: "/workloads",
              subNav: [
                {
                  title: "VirtualMachines",
                  itemId: "/workloads/vms"
                  // Optional
                  // elemBefore: () => <Icon name="calendar" />
                },
                {
                  title: "VirtualMachinesInstances",
                  itemId: "/workloads/vmis"
                  // Optional
                  // elemBefore: () => <Icon name="calendar" />
                },
                {
                  title: "Migrations",
                  itemId: "/workloads/vmims"
                  // Optional
                  // elemBefore: () => <Icon name="calendar" />
                },
                {
                  title: "Pods",
                  itemId: "/workloads/pods"
                  // elemBefore: () => <Icon name="calendar" />
                },
              ]
            },
            {
              title: "Nodes",
              itemId: "/nodes",
            },
          ]}
        />

        <div className="absolute bottom-0 w-full my-8">
          <Navigation
            activeItemId={location.pathname}
            items={[
              {
                title: "Settings",
                itemId: "/settings",
                elemBefore: () => <Icon name="settings" />
              }
            ]}
            onSelect={({ itemId }) => {
              history.push(itemId);
            }}
          />
        </div>
      </div>
    </React.Fragment>
  );
};
