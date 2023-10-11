import * as React from 'react';
import { Link } from 'react-router-dom';
import axios from 'axios';
import {
  Bullseye,
  Spinner,
  Card,
  CardTitle,
  CardBody,
  Button,
  List,
  Modal,
  ModalVariant,
  ListItem,
  Chip,
  Grid,
  GridItem,
  CardFooter,
  Icon,
  gridSpans,
  Tooltip,
  Page,
  PageSection,
  Title,
  CodeBlock,
  CodeBlockCode, SimpleList, SimpleListGroup, SimpleListItem,
} from "@patternfly/react-core";
import {apiBaseUrl} from "@app/config";
import {CheckCircleIcon, ExclamationCircleIcon, ExclamationTriangleIcon} from "@patternfly/react-icons";
import {Table, TableHeader, TableBody, TableProps, TableVariant} from '@patternfly/react-table';

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
  const [importedMustGathers, setImportedMustGathers] = React.useState<data | undefined>(undefined);
  const [mustGatherInsightData, setMustGatherInsightData] = React.useState<any | undefined>(undefined);

  React.useEffect(() => {
    fetch("/getResourceStats", setResourceCount)
    fetch("/nodes", setNodes)
    fetch("/vmis", setVmis)
    fetch("/pods", setPods)
    fetch("/getSubscriptions", setSubscriptions)
    fetch("/getImportedMustGathers", setImportedMustGathers)
  }, []);

  const fetch = (path, callback) => {
    axios
      .get(apiBaseUrl + path + "?per_page=5")
      .then((response) => {
        console.log(response.data);
        callback(response.data);
      });
  }

  const gridItems = (title, link, counts) => {
    const style = {
      marginLeft: "auto",
      marginRight: "5px",
    }

    const gridSpan: gridSpans | undefined = 12-((counts.slice(1).filter((count) => count > 0).length) * 2) as gridSpans

    return <Grid hasGutter>
      <GridItem span={gridSpan}><Link to={link}>{counts[0]} {title}</Link></GridItem>
      {
        counts.length > 1 && counts[1] > 0 &&
          <GridItem span={2} style={style}>
            <Link to={link+"?status=healthy"}>
              {counts[1]}
                <Icon status="success" className="pf-u-ml-xs">
                    <CheckCircleIcon />
                </Icon>
            </Link>
          </GridItem>
      }
      {
        counts.length > 2 && counts[2] > 0 &&
          <GridItem span={2} style={style}>
            <Link to={link+"?status=unhealthy"}>
              {counts[2]}
              <Icon status="danger" className="pf-u-ml-xs">
                  <ExclamationCircleIcon />
              </Icon>
            </Link>
          </GridItem>
      }
      {
        counts.length > 3 && counts[3] > 0 &&
          <GridItem span={2} style={style}>
            <Link to={link+"?status=warning"}>
              {counts[3]}
              <Icon status="warning" className="pf-u-ml-xs">
                <ExclamationTriangleIcon />
              </Icon>
            </Link>
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
      "Nodes", "/nodes",
      [nodesCount, readyNodesCount, notReadyNodesCount],
    );

    const vmisGrid = gridItems(
      "VMIs", "/workloads/virtualmachineinstances",
      [vmisCount, healthyVmisCount, failedVmisCount, otherVmisCount],
    );

    const vmimsGrid = gridItems(
      "VMI Migrations", "/workloads/migrations",
      [vmimsCount, healthyVmimsCount, failedVmimsCount, otherVmimsCount],
    );

    const podsGrid = gridItems(
      "Pods", "/workloads/pods",
      [podsCount, healthyPodsCount, failedPodsCount, otherPodsCount],
    );

    const pvcsGrid = gridItems(
      "PVCs", "/storage/pvcs",
      [pvcsCount, healthyPvcsCount, failedPvcsCount, otherPvcsCount],
    );

    return (
      <Card style={{ height: "100%" }}>
        <CardTitle>Cluster inventory</CardTitle>
        <CardBody>
          <List isPlain isBordered>
            <ListItem>{nodesGrid}</ListItem>
            <ListItem>{vmisGrid}</ListItem>
            <ListItem>{vmimsGrid}</ListItem>
            <ListItem>{podsGrid}</ListItem>
            <ListItem>{pvcsGrid}</ListItem>
          </List>
        </CardBody>
      </Card>
    )
  }

  const nodesStatusCard = () => {
    return (
      <Card style={{ height: "100%" }}>
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
          <Link to="/nodes">
            <Button variant="link" isInline component="a">
              View more
            </Button>
          </Link>
        </CardFooter>
      </Card>
    )
  }

  const virtualMachineInstancesCard = () => {
    return (
      <Card style={{ height: "100%" }}>
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
          <Link to="/workloads/virtualmachineinstances">
            <Button variant="link" isInline component="a">
              View more
            </Button>
          </Link>
        </CardFooter>
      </Card>
    )
  }

  const podsCard = () => {
    return (
      <Card style={{ height: "100%" }}>
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
          <Link to="/workloads/pods">
            <Button variant="link" isInline component="a">
              View more
            </Button>
          </Link>
        </CardFooter>
      </Card>
    )
  }

  const subscriptionsCard = () => {
    return (
      <Card style={{ height: "100%" }}>
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
          <Link to="/subscriptions">
            <Button variant="link" isInline component="a">
              View more
            </Button>
          </Link>
        </CardFooter>
      </Card>
    )
  }

  const importedMustGathersCard = () => {
    const columnNames = {
      name: 'File name',
      importTime: 'Import date',
      gatherTime: 'Gather date',
    };

    const columns: TableProps['cells'] = ['File name', 'Gather date', 'Import date'];

    let rows: TableProps['rows'] = [
      ['Loading...', 'Loading...', 'Loading...'],
    ];

    if (importedMustGathers !== undefined) {
      if (importedMustGathers.data.length === 0) {
        rows = [
          ['No must-gathers imported', '', ''],
        ];
      }

      rows = importedMustGathers.data.map((mustGather) => [
        <span onClick={() => setMustGatherInsightData(mustGather.insightsData)}>{mustGather.name}</span>,
        <span onClick={() => setMustGatherInsightData(mustGather.insightsData)}>{new Date(mustGather.gatherTime).toLocaleString()}</span>,
        <span onClick={() => setMustGatherInsightData(mustGather.insightsData)}>{new Date(mustGather.importTime).toLocaleString()}</span>,
      ]);
    }

    return (
      <Card style={{ height: "100%" }}>
        <CardTitle>Imported Must-Gathers</CardTitle>
        <CardBody>
          {
            importedMustGathers === undefined ? (
              <Bullseye>
                <Spinner aria-label="Loading data" />
              </Bullseye>
            ) : (
              <Table
                variant={TableVariant.compact}
                cells={columns}
                rows={rows}
              >
                <TableHeader />
                <TableBody />
              </Table>
            )
          }
        </CardBody>
      </Card>
    )
  }

  const reportTable = (report) => {
    return (
      <div className="pf-u-mt-xl">
        <Title headingLevel="h2" size="md">{report["key"]} ({report["type"]} - {report["component"]})</Title>

        <CodeBlock className="pf-u-mt-sm">
          <CodeBlockCode id="code-content">{JSON.stringify(report, null, 2)}</CodeBlockCode>
        </CodeBlock>
      </div>
    )
  }

  const mustGatherInsightsDataModalContent = (data) => {
    const metadata = data["analysis_metadata"]
    const plugins = metadata["plugin_sets"]

    return (
      <>
        <SimpleList isControlled={false} className="pf-u-mb-xl">
          <SimpleListGroup title="Analysis Metadata">
            <SimpleListItem><b>Start:</b> {new Date(metadata["start"]).toLocaleString()}</SimpleListItem>
            <SimpleListItem><b>Finish:</b> {new Date(metadata["finish"]).toLocaleString()}</SimpleListItem>
            <SimpleListItem><b>Execution Context:</b> {metadata["execution_context"]}</SimpleListItem>
          </SimpleListGroup>
          <SimpleListGroup title="Plugin Sets">
            <SimpleListItem><b>insights-core:</b> {plugins["insights-core"]["version"]}</SimpleListItem>
            <SimpleListItem><b>ccx_rules_ocp:</b> {plugins["ccx_rules_ocp"]["version"]}</SimpleListItem>
            <SimpleListItem><b>ccx_ocp_core:</b> {plugins["ccx_ocp_core"]["version"]}</SimpleListItem>
          </SimpleListGroup>
        </SimpleList>

        {data["reports"].map((report) => reportTable(report))}
      </>
    )
  }

  const mustGatherInsightsDataModal = () => {
    return (
      <Modal
        variant={ModalVariant.medium}
        title="Insights data"
        isOpen={mustGatherInsightData !== undefined && mustGatherInsightData != null}
        onClose={() => {
          setMustGatherInsightData(undefined)
        }}
        actions={[
          <Button variant="link" onClick={() => {
            setMustGatherInsightData(undefined)
          }}>
            Close
          </Button>
        ]}
      >
        {(mustGatherInsightData !== undefined && mustGatherInsightData !== null) ?
          mustGatherInsightsDataModalContent(JSON.parse(mustGatherInsightData))
          :
          <span>No insights data found</span>
        }
      </Modal>
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
        {mustGatherInsightsDataModal()}

        <Grid hasGutter>
          {/* line 1 */}
          <GridItem span={6}>{clusterInventoryCard()}</GridItem>
          <GridItem span={6}>{importedMustGathersCard()}</GridItem>

          {/* line 2 */}
          <GridItem span={6}>{subscriptionsCard()}</GridItem>
          <GridItem span={6}>{virtualMachineInstancesCard()}</GridItem>

          {/* line 3 */}
          <GridItem span={6}>{podsCard()}</GridItem>
          <GridItem span={6}>{nodesStatusCard()}</GridItem>
        </Grid>
      </PageSection>
    </Page>
  )
}

export { Dashboard };
