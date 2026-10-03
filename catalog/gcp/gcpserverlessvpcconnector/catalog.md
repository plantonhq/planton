# GCP Serverless VPC Connector

Deploys a Serverless VPC Access connector — the managed bridge that lets serverless workloads (Cloud Run, Cloud Functions, App Engine) send traffic into a VPC network. The connector runs a small fleet of forwarding instances inside the network; serverless egress configured to use it reaches private IPs (Cloud SQL private IP, Memorystore, internal load balancers) as if the workload lived in the VPC. One connector serves many services in its region — it is shared infrastructure, not per-workload.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **VPC Access API enablement** on the target project (never disabled on destroy)
- **Serverless VPC Access Connector** -- a `google_vpc_access_connector` occupying a dedicated /28: either carved directly out of the chosen VPC (network placement) or an existing /28 subnetwork (subnet placement — the Shared VPC shape), with the configured machine type and scaling window

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the connector will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **IAM permissions** -- `roles/vpcaccess.admin` (or equivalent) on the target project.
- **Address space** -- an unused /28 in the target VPC (network placement), or a dedicated /28 GcpSubnetwork (subnet placement; on Shared VPC it lives in the host project).

## Deploy

### Console

Open the deployment store, find **GCP Serverless VPC Connector**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Private egress — basic** preset in the [Presets](#presets) tab for the common single-project shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpServerlessVpcConnector
metadata:
  name: svcless-uc1
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  region: us-central1
  network:
    valueFrom:
      kind: GcpVpcNetwork
      name: prod-vpc
      fieldPath: status.outputs.network_name
  ipCidrRange: 10.8.0.0/28
  machineType: e2-micro
```

```shell
planton apply -f connector.yaml
```

This carves a dedicated /28 out of the VPC and stands up the forwarding fleet; Cloud Run services then attach by the connector's full resource name. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the connector's placement to network resources deployed in the same InfraPipeline — here the Shared VPC shape, occupying a dedicated /28 subnetwork:

```yaml
spec:
  region: us-central1
  subnet:
    name:
      valueFrom:
        kind: GcpSubnetwork
        name: connector-slash28
        fieldPath: status.outputs.subnetwork_name
```

The InfraPipeline deploys the subnetwork first, then the connector against it — and serverless consumers downstream reference this connector's `self_link`. The classic chain: GcpVpcNetwork → GcpSubnetwork → GcpCloudSql (private IP) + GcpServerlessVpcConnector → GcpCloudRun, all composed by reference.

## Key Configuration

These are the most important decisions when configuring a connector. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Placement (exactly one)** -- `network` + `ipCidrRange` carves a new /28 out of the VPC (the range must not overlap any subnet, peered range, or route). `subnet` occupies an existing /28 subnetwork created for the connector — REQUIRED on Shared VPC, where the range lives in the host project (`subnet.projectId`). GCP demands exactly a /28. The whole placement is create-time.

**Region** -- Serverless workloads can only use a connector in their OWN region; multi-region services need one per region. Immutable.

**Machine type** -- Per-instance throughput class: `f1-micro` (~100 Mbps), `e2-micro` (~200 Mbps, the recommended default), `e2-standard-4` (~1 Gbps class). Mutable in place — the safe capacity lever.

**Scaling window** -- `minInstances` (2-9) strictly below `maxInstances` (3-10). Two sharp edges: the fleet NEVER scales in on its own (it holds the post-burst high-water mark until manually reduced), and DECREASING either value replaces the connector — a brief egress outage. Increases apply in place.

**Teardown blast radius** -- Deleting the connector stops private egress for EVERY function and service attached to it in the region (the fleet tears down over ~3-5 minutes). Set `deletionPolicy: PREVENT` on shared production connectors so a destroy fails instead; `ABANDON` unmanages the connector while it keeps forwarding.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** (network placement) | `network` | `status.outputs.network_name` |
| **GcpSubnetwork** (subnet placement) | `subnet.name` | `status.outputs.subnetwork_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `self_link` | Full resource name (projects/*/locations/*/connectors/*) | GcpCloudRun VPC access, GcpCloudFunction connector setting — the attachment handle |
| `name` | Short name of the connector | Monitoring, gcloud runbooks |
| `state` | READY / CREATING / DELETING / ERROR / UPDATING | Health checks — a non-READY state explains failing egress |
| `region` | The plain region name | Scope-compatibility checks in compositions |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Private Egress Basic** -- The single-project default: carve a /28, e2-micro, GCP's 2/10 scaling. Start from the **Private egress — basic** preset.

**High Throughput** -- e2-standard-4 instances with a widened scaling floor for data-heavy serverless (bulk writes, media pipelines). Start from the **High throughput — production** preset.

**Shared VPC Subnet** -- Subnet placement against a host-project /28 — the shape platform teams publish for service projects. Start from the **Shared VPC — subnet placement** preset.

## Works With

- [**GCP Cloud Run**](/infra-catalog/gcp-cloud-run) -- attaches by the connector's resource name to reach private IPs
- [**GCP Cloud Function**](/infra-catalog/gcp-cloud-function) -- same attachment for Cloud Functions egress
- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) -- the network the connector bridges into
- [**GCP Subnetwork**](/infra-catalog/gcp-subnetwork) -- provides the dedicated /28 in subnet placement (Shared VPC)
- [**GCP Cloud SQL**](/infra-catalog/gcp-cloud-sql) -- the classic private-IP destination behind the connector
- [**GCP Memorystore Instance**](/infra-catalog/gcp-memorystore-instance) -- private Redis reached through the connector
- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the connector is created
