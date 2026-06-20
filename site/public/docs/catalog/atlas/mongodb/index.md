---
title: "MongoDB"
description: "MongoDB deployment documentation"
icon: "package"
order: 100
componentName: "atlasmongodb"
---

# Atlas MongoDB

Deploys a Atlas MongoDB cluster with configurable cluster type, replication topology, cloud provider selection, instance sizing, and backup settings. All cluster-level parameters are managed declaratively through the spec, including electable and read-only node counts, auto-scaling, and MongoDB version.

## What Gets Created

When you deploy a AtlasMongodb resource, OpenMCF provisions:

- **Atlas MongoDB Advanced Cluster** — a `atlasmongodb_advanced_cluster` resource with the specified cluster type, replication spec, region configuration, electable/read-only node topology, backup, and auto-scaling settings
- **Atlas MongoDB Provider** — configured using explicit API key credentials from provider config or environment variables

## Prerequisites

- **Atlas MongoDB API keys** configured via environment variables (`MONGODB_ATLAS_PUBLIC_KEY`, `MONGODB_ATLAS_PRIVATE_KEY`) or OpenMCF provider config
- **A Atlas MongoDB project** with its project ID — the cluster is created inside this project
- **Sufficient Atlas permissions** for the API key to create and manage clusters in the target project

## Quick Start

Create a file `atlas-mongodb.yaml`:

```yaml
apiVersion: atlas.openmcf.org/v1
kind: AtlasMongodb
metadata:
  name: my-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AtlasMongodb.my-cluster
spec:
  clusterConfig:
    projectId: "64a1234567890abcdef12345"
    clusterType: REPLICASET
    electableNodes: 3
    priority: 7
    providerName: AWS
    providerInstanceSizeName: M10
    mongoDbMajorVersion: "7.0"
```

Deploy:

```shell
openmcf apply -f atlas-mongodb.yaml
```

This creates a 3-node replica set cluster on AWS using M10 instances running MongoDB 7.0.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `clusterConfig.projectId` | `string` | The unique ID for the Atlas MongoDB project where the cluster is created. | — |
| `clusterConfig.clusterType` | `string` | Type of cluster to deploy. Accepted values: `REPLICASET`, `SHARDED`, `GEOSHARDED`. You cannot convert a sharded cluster to a replica set after creation. | — |
| `clusterConfig.electableNodes` | `int32` | Number of electable nodes in the region. Electable nodes can become primary and serve local reads. Total across all regions must be 3, 5, or 7. Set to `0` if the region has `priority` of `0`. | — |
| `clusterConfig.priority` | `int32` | Election priority of the region (1-7). Priority `7` designates the Preferred Region where Atlas places the primary node. Each region must have a unique priority, with each region exactly one less than the previous. | — |
| `clusterConfig.providerName` | `string` | Cloud provider for the cluster servers. Accepted values: `AWS`, `GCP`, `AZURE`, `TENANT` (multi-tenant, only valid with M2 or M5 instance sizes). | — |
| `clusterConfig.providerInstanceSizeName` | `string` | Instance size for all data-bearing servers in the cluster. Each size has a default storage capacity and RAM allocation. Examples: `M0`, `M2`, `M5`, `M10`, `M20`, `M30`, `M40`, `M50`, `M60`, `M80`, `M200`, `M300`. | — |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `clusterConfig.readOnlyNodes` | `int32` | `0` | Number of read-only nodes in the region. Read-only nodes can never become primary but can serve local reads. |
| `clusterConfig.cloudBackup` | `bool` | `false` | Enable or disable cloud backup for the cluster. |
| `clusterConfig.autoScalingDiskGbEnabled` | `bool` | `false` | Enable automatic disk storage scaling. When enabled, Atlas automatically increases storage capacity when usage approaches the provisioned limit. |
| `clusterConfig.mongoDbMajorVersion` | `string` | `"7.0"` | MongoDB major version to deploy. Supported versions for M10+ clusters: `4.4`, `5.0`, `6.0`, `7.0`. If omitted, Atlas deploys 7.0. For M0, M2, or M5 instances, Atlas deploys 5.0. |

## Examples

### Basic Replica Set on AWS

A minimal 3-node replica set on AWS with M10 instances, suitable for development or small production workloads:

```yaml
apiVersion: atlas.openmcf.org/v1
kind: AtlasMongodb
metadata:
  name: dev-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AtlasMongodb.dev-cluster
spec:
  clusterConfig:
    projectId: "64a1234567890abcdef12345"
    clusterType: REPLICASET
    electableNodes: 3
    priority: 7
    providerName: AWS
    providerInstanceSizeName: M10
    mongoDbMajorVersion: "7.0"
```

### Production Replica Set with Backup and Auto-Scaling

A production-grade replica set on GCP with cloud backup enabled, auto-scaling disk storage, and read-only nodes for offloading analytics queries:

```yaml
apiVersion: atlas.openmcf.org/v1
kind: AtlasMongodb
metadata:
  name: prod-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AtlasMongodb.prod-cluster
spec:
  clusterConfig:
    projectId: "64a1234567890abcdef12345"
    clusterType: REPLICASET
    electableNodes: 3
    priority: 7
    readOnlyNodes: 2
    cloudBackup: true
    autoScalingDiskGbEnabled: true
    providerName: GCP
    providerInstanceSizeName: M30
    mongoDbMajorVersion: "7.0"
```

### Sharded Cluster on Azure

A sharded cluster on Azure with M40 instances for high-throughput workloads that require horizontal scaling across multiple shards:

```yaml
apiVersion: atlas.openmcf.org/v1
kind: AtlasMongodb
metadata:
  name: analytics-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AtlasMongodb.analytics-cluster
spec:
  clusterConfig:
    projectId: "64a1234567890abcdef12345"
    clusterType: SHARDED
    electableNodes: 3
    priority: 7
    readOnlyNodes: 1
    cloudBackup: true
    autoScalingDiskGbEnabled: true
    providerName: AZURE
    providerInstanceSizeName: M40
    mongoDbMajorVersion: "6.0"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `id` | `string` | Provider-assigned unique ID for the Atlas MongoDB cluster |
| `bootstrapEndpoint` | `string` | Standard connection string in SRV format (`mongodb+srv://...`), recommended for MongoDB drivers |
| `crn` | `string` | The cluster identifier, used for resource identification and API operations |
| `restEndpoint` | `string` | Standard connection string in legacy format (`mongodb://host:port,...`) |

## Related Components

No other OpenMCF components have direct foreign key references to AtlasMongodb. This component is typically deployed as a standalone resource within a Atlas MongoDB project.
