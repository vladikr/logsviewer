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
  ListItem, Page, Chip, Grid, GridItem, CardFooter,
} from "@patternfly/react-core";
import {apiBaseUrl} from "@app/config";

type data = {
  meta: {
    totalRowCount: number;
  }
  data: any[];
}

const Dashboard: React.FunctionComponent = () => {
  const [nodes, setNodes] = React.useState<data | undefined>(undefined);
  const [vmis, setVmis] = React.useState<data | undefined>(undefined);
  const [pods, setPods] = React.useState<data | undefined>(undefined);
  const [pvcs, setPvcs] = React.useState<data | undefined>(undefined);

  React.useEffect(() => {
    fetch("/nodes", setNodes)
    fetch("/vmis", setVmis)
    fetch("/pods", setPods)
    fetch("/getPVCs", setPvcs)
  }, []);

  const fetch = (path, callback) => {
    axios
      .get(apiBaseUrl + path + "?per_page=5")
      .then((response) => {
        console.log(response.data);
        callback(response.data);
      });
  }

  const clusterInventoryCard = () => {
    return (
      <Card>
        <CardTitle>Cluster inventory</CardTitle>
        <CardBody>
          {
            nodes === undefined || vmis === undefined || pods === undefined || pvcs === undefined ? (
              <Bullseye>
                <Spinner aria-label="Loading data" />
              </Bullseye>
            ) : (
              <List isPlain isBordered>
                <ListItem><a href="/nodes">
                  {nodes.meta.totalRowCount} Nodes
                </a></ListItem>
                <ListItem><a href="/workloads/virtualmachineinstances">
                  {vmis.meta.totalRowCount} Virtual Machine Instances
                </a></ListItem>
                <ListItem><a href="/workloads/pods">
                  {pods.meta.totalRowCount} Pods
                </a></ListItem>
                <ListItem><a href="/storage/pvcs">
                  {pvcs.meta.totalRowCount} PVCs
                </a></ListItem>
              </List>
            )
          }
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

  return (
    <Page isManagedSidebar>
      <PageSection>
        <Title headingLevel="h1" size="4xl">
          Overview
        </Title>
      </PageSection>

      <PageSection>
        <Grid hasGutter>
          <GridItem span={4} style={{ height: "100%" }}>{clusterInventoryCard()}</GridItem>
          <GridItem span={8}>{nodesStatusCard()}</GridItem>
          <GridItem span={12}>{virtualMachineInstancesCard()}</GridItem>
          <GridItem span={12}>{podsCard()}</GridItem>
        </Grid>
      </PageSection>
    </Page>
  )
}

export { Dashboard };
