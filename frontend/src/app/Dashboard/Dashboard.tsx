import * as React from 'react';
import axios from 'axios';
import {
  PageSection,
  Bullseye,
  Spinner,
  Title,
  Card,
  CardTitle,
  CardBody,
  Button,
  List,
  ListItem, Page, Chip, Grid, GridItem, CardFooter, Icon, Flex, gridSpans, Tooltip,
} from "@patternfly/react-core";
import {apiBaseUrl} from "@app/config";
import {CheckCircleIcon, ExclamationCircleIcon, ExclamationTriangleIcon} from "@patternfly/react-icons";

type data = {
  meta: {
    totalRowCount: number;
  }
  data: any[];
}

type resourceStatus = {
  [status: string]: number;
}

type resourceCount = {
  nodes: resourceStatus;
  pods: resourceStatus;
  pvcs: resourceStatus;
  vmimigrations: resourceStatus;
  vmis: resourceStatus;
}

const Dashboard: React.FunctionComponent = () => {
  const [resourceCount, setResourceCount] = React.useState<resourceCount>({
    nodes: {},
    pods: {},
    pvcs: {},
    vmimigrations: {},
    vmis: {}
  });
  const [nodes, setNodes] = React.useState<data | undefined>(undefined);
  const [vmis, setVmis] = React.useState<data | undefined>(undefined);
  const [pods, setPods] = React.useState<data | undefined>(undefined);
  const [subscriptions, setSubscriptions] = React.useState<data | undefined>(undefined);

  React.useEffect(() => {
    fetch("/getResourceStats", setResourceCount)
    fetch("/nodes", setNodes)
    fetch("/vmis", setVmis)
    fetch("/pods", setPods)
    fetch("/getSubscriptions", setSubscriptions)
  }, []);

  const fetch = (path, callback) => {
    axios
      .get(apiBaseUrl + path + "?per_page=5")
      .then((response) => {
        console.log(response.data);
        callback(response.data);
      });
  }

  const gridItems = (title, counts) => {
    const style = {
      marginLeft: "auto",
      marginRight: "5px",
    }

    const gridSpan: gridSpans | undefined = 12-((counts.slice(1).filter((count) => count > 0).length) * 2) as gridSpans

    return <Grid hasGutter>
      <GridItem span={gridSpan}>{counts[0]} {title}</GridItem>
      {
        counts.length > 1 && counts[1] > 0 &&
          <GridItem span={2} style={style}>
            {counts[1]}
              <Icon status="success" className="pf-u-ml-xs">
                  <CheckCircleIcon />
              </Icon>
          </GridItem>
      }
      {
        counts.length > 2 && counts[2] > 0 &&
          <GridItem span={2} style={style}>
            {counts[2]}
              <Icon status="danger" className="pf-u-ml-xs">
                  <ExclamationCircleIcon />
              </Icon>
          </GridItem>
      }
      {
        counts.length > 3 && counts[3] > 0 &&
          <GridItem span={2} style={style}>
            {counts[3]}
              <Icon status="warning" className="pf-u-ml-xs">
                <ExclamationTriangleIcon />
              </Icon>
          </GridItem>
      }
    </Grid>
  }

  const clusterInventoryCard = () => {
    let nodesCount = 0, readyNodesCount = 0, notReadyNodesCount = 0,
      vmisCount = 0, healthyVmisCount = 0, otherVmisCount = 0, failedVmisCount = 0,
      vmimsCount = 0, healthyVmimsCount = 0, otherVmimsCount = 0, failedVmimsCount = 0,
      podsCount = 0, healthyPodsCount = 0, failedPodsCount = 0, otherPodsCount = 0,
      pvcsCount = 0, healthyPvcsCount = 0, failedPvcsCount = 0, otherPvcsCount = 0;

    for (let nodesKey in resourceCount.nodes) {
      nodesCount += resourceCount.nodes[nodesKey];
      if (nodesKey === "Ready") {
        readyNodesCount += resourceCount.nodes[nodesKey];
      } else {
        notReadyNodesCount += resourceCount.nodes[nodesKey];
      }
    }

    for (let vmisKey in resourceCount.vmis) {
      vmisCount += resourceCount.vmis[vmisKey];
      if (vmisKey === "Running" || vmisKey === "Succeeded") {
        healthyVmisCount += resourceCount.vmis[vmisKey];
      } else if (vmisKey === "Failed") {
        failedVmisCount += resourceCount.vmis[vmisKey];
      } else {
        otherVmisCount += resourceCount.vmis[vmisKey];
      }
    }
    for (let vmimsKey in resourceCount.vmimigrations) {
      vmimsCount += resourceCount.vmimigrations[vmimsKey];
      if (vmimsKey === "Running" || vmimsKey === "Succeeded") {
        healthyVmimsCount += resourceCount.vmimigrations[vmimsKey];
      } else if (vmimsKey === "Failed") {
        failedVmimsCount += resourceCount.vmimigrations[vmimsKey];
      } else {
        otherVmimsCount += resourceCount.vmimigrations[vmimsKey];
      }
    }
    for (let podsKey in resourceCount.pods) {
      podsCount += resourceCount.pods[podsKey];
      if (podsKey === "Running" || podsKey === "Succeeded") {
        healthyPodsCount += resourceCount.pods[podsKey];
      } else if (podsKey === "Failed") {
        failedPodsCount += resourceCount.pods[podsKey];
      } else {
        otherPodsCount += resourceCount.pods[podsKey];
      }
    }

    for (let pvcsKey in resourceCount.pvcs) {
      pvcsCount += resourceCount.pvcs[pvcsKey];
      if (pvcsKey === "Bound") {
        healthyPvcsCount += resourceCount.pvcs[pvcsKey];
      } else if (pvcsKey === "Lost") {
        failedPvcsCount += resourceCount.pvcs[pvcsKey];
      } else {
        otherPvcsCount += resourceCount.pvcs[pvcsKey];
      }
    }

    const nodesGrid = gridItems(
      "Nodes",
      [nodesCount, readyNodesCount, notReadyNodesCount],
    );

    const vmisGrid = gridItems(
      "VMIs",
      [vmisCount, healthyVmisCount, failedVmisCount, otherVmisCount],
    );

    const vmimsGrid = gridItems(
      "VMI Migrations",
      [vmimsCount, healthyVmimsCount, failedVmimsCount, otherVmimsCount],
    );

    const podsGrid = gridItems(
      "Pods",
      [podsCount, healthyPodsCount, failedPodsCount, otherPodsCount],
    );

    const pvcsGrid = gridItems(
      "PVCs",
      [pvcsCount, healthyPvcsCount, failedPvcsCount, otherPvcsCount],
    );

    return (
      <Card>
        <CardTitle>Cluster inventory</CardTitle>
        <CardBody>
          <List isPlain isBordered>
            <ListItem><a href="/nodes">{nodesGrid}</a></ListItem>
            <ListItem><a href="/workloads/virtualmachineinstances">{vmisGrid}</a></ListItem>
            <ListItem><a href="/workloads/migrations">{vmimsGrid}</a></ListItem>
            <ListItem><a href="/workloads/pods">{podsGrid}</a></ListItem>
            <ListItem><a href="/storage/pvcs">{pvcsGrid}</a></ListItem>
          </List>
        </CardBody>
      </Card>
    )
  }

  const nodesStatusCard = () => {
    return (
      <Card>
        <CardTitle>Nodes Status</CardTitle>
        <CardBody>
          {
            nodes === undefined ? (
              <Bullseye>
                <Spinner aria-label="Loading data" />
              </Bullseye>
            ) : (
              <List isPlain>
                {
                  nodes.data.length === 0 ? (
                    <ListItem>No nodes found</ListItem>
                  ) :
                  nodes.data.slice(0, 5).map((node) => {
                    return (
                      <ListItem key={node.systemUuid}>
                        <Chip isReadOnly
                          style={{
                            marginRight: "5px",
                            backgroundColor: node.status === "Ready" ? "var(--pf-global--palette--green-200)" : "var(--pf-global--palette--red-200)",
                          }}
                        >{node.status}</Chip>
                        {node.name}
                      </ListItem>
                    )
                  })
                }
              </List>
            )
          }
        </CardBody>
        <CardFooter>
          <Button variant="link" isInline component="a" href="/nodes">
            View more
          </Button>
        </CardFooter>
      </Card>
    )
  }

  const virtualMachineInstancesCard = () => {
    return (
      <Card>
        <CardTitle>VMIs</CardTitle>
        <CardBody>
          {
            vmis === undefined ? (
              <Bullseye>
                <Spinner aria-label="Loading data" />
              </Bullseye>
            ) : (
              <List isPlain>
                {
                  vmis.data.length === 0 ? (
                    <ListItem>No VMIs found</ListItem>
                  ) :
                  vmis.data.slice(0, 5).map((vmi) => {
                    return (
                      <ListItem key={vmi.uuid}>
                        <Chip isReadOnly
                              style={{
                                marginRight: "5px",
                              }}
                        >{vmi.phase}</Chip>
                        {`${vmi.namespace}/${vmi.name}`}
                      </ListItem>
                    )
                  })
                }
              </List>
            )
          }
        </CardBody>
        <CardFooter>
          <Button variant="link" isInline component="a" href="/workloads/virtualmachineinstances">
            View more
          </Button>
        </CardFooter>
      </Card>
    )
  }

  const podsCard = () => {
    return (
      <Card>
        <CardTitle>Pods</CardTitle>
        <CardBody>
          {
            pods === undefined ? (
              <Bullseye>
                <Spinner aria-label="Loading data" />
              </Bullseye>
            ) : (
              <List isPlain>
                {
                  pods.data.length === 0 ? (
                    <ListItem>No pods found</ListItem>
                  ) :
                  pods.data.slice(0, 5).map((pod) => {
                    return (
                      <ListItem key={pod.uuid}>
                        <Chip isReadOnly
                              style={{
                                marginRight: "5px",
                              }}
                        >{pod.phase}</Chip>
                        {`${pod.namespace}/${pod.name}`}
                      </ListItem>
                    )
                  })
                }
              </List>
            )
          }
        </CardBody>
        <CardFooter>
          <Button variant="link" isInline component="a" href="/workloads/pods">
            View more
          </Button>
        </CardFooter>
      </Card>
    )
  }

  const subscriptionsCard = () => {
    return (
      <Card>
        <CardTitle>Installed Operators</CardTitle>
        <CardBody>
          {
            subscriptions === undefined ? (
              <Bullseye>
                <Spinner aria-label="Loading data" />
              </Bullseye>
            ) : (
              <List isPlain>
                {
                  subscriptions.data.length === 0 ? (
                    <ListItem>No operators installed</ListItem>
                  ) :
                  subscriptions.data.slice(0, 5).map((subscription) => {
                    return (
                      <ListItem key={subscription.uuid}>
                        <Tooltip
                          content={
                            <div>
                              <p>installed CSV: {subscription.installedCSV}</p>
                              <p>source: {subscription.sourceNamespace}/{subscription.source}</p>
                            </div>
                          }
                        >
                          <div>
                            <Chip isReadOnly
                                  style={{
                                    marginRight: "5px",
                                  }}
                            >{subscription.state}</Chip>
                            {subscription.namespace}/{subscription.name}
                          </div>
                        </Tooltip>
                      </ListItem>
                    )
                  })
                }
              </List>
            )
          }
        </CardBody>
        <CardFooter>
          <Button variant="link" isInline component="a" href="/subscriptions">
            View more
          </Button>
        </CardFooter>
      </Card>
    )
  }

  return (
    <Page isManagedSidebar>
      <PageSection>
        <Title headingLevel="h1" size="4xl">
          Overview
        </Title>
      </PageSection>

      <PageSection>
        <Grid hasGutter>
          {/* line 1 */}
          <GridItem span={6}>{clusterInventoryCard()}</GridItem>
          <GridItem span={6}>{subscriptionsCard()}</GridItem>

          {/* line 2 */}
          <GridItem span={12}>{virtualMachineInstancesCard()}</GridItem>

          {/* line 3 */}
          <GridItem span={6}>{podsCard()}</GridItem>
          <GridItem span={6}>{nodesStatusCard()}</GridItem>
        </Grid>
      </PageSection>
    </Page>
  )
}

export { Dashboard };
