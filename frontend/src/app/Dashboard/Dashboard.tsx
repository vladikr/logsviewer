import * as React from 'react';
import axios from 'axios';
import {
  PageSection,
  Bullseye, EmptyState, EmptyStateIcon, Spinner, Title,
} from "@patternfly/react-core";

const Dashboard: React.FunctionComponent = () => {

    return (
      <PageSection>
        <Title headingLevel="h1" size="lg">Dashboard Page TBD!</Title>

        {/* must-gather.local./event-filter.html */}
        <div>
          List of events
        </div>

        {/* must-gather.local./timestamp */}
        <div>
          Must-gather timestamp - 2023-05-04 17:28:14.902535465
        </div>
      </PageSection>
    )
}

export { Dashboard };
