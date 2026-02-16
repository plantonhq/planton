---
title: "Bigtable Instance"
description: "Bigtable Instance deployment documentation"
icon: "package"
order: 100
componentName: "gcpbigtableinstance"
---

# GCP Bigtable Instance

Deploys a Cloud Bigtable instance with one or more clusters, supporting SSD and HDD storage types, per-cluster autoscaling, CMEK encryption, and multi-cluster replication. Tables and app profiles are application-level concerns managed separately.

## What Gets Created

When you deploy a GcpBigtableInstance resource, OpenMCF provisions:

- **Bigtable Instance** — a `google_bigtable_instance` resource that serves as the logical container for data, with GCP labels applied automatically
- **One or more Clusters** — inline cluster configurations within the instance, each placed in a specific zone with independent scaling (fixed or autoscaling) and storage type settings
- **CMEK Encryption** — created only when `kmsKeyName` is set on a cluster, encrypts data at rest using the specified Cloud KMS key

## Prerequisites

- **GCP credentials** configured via environment variables or OpenMCF provider config
- **A GCP project** where the Bigtable instance will be created
- **Zones** that support Bigtable instances (see [GCP Bigtable locations](https://cloud.google.com/bigtable/docs/locations))
- **A Cloud KMS key** if enabling CMEK encryption (key region must match cluster zone region)

## Quick Start

Create a file `bigtable.yaml`:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpBigtableInstance
metadata:
  name: my-bigtable
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpBigtableInstance.my-bigtable
spec:
  projectId:
    value: my-gcp-project
  instanceName: my-bigtable-instance
  clusters:
    - clusterId: my-cluster-01
      zone: us-central1-a
```

Deploy:

```shell
openmcf apply -f bigtable.yaml
```

This creates a Bigtable instance with a single SSD cluster in `us-central1-a`. Bigtable auto-allocates nodes based on data footprint since neither `numNodes` nor `autoscalingConfig` is specified.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `string` | GCP project ID. Can reference a GcpProject resource via `valueFrom`. | Required |
| `instanceName` | `string` | Instance name (also the Instance ID in GCP Console). | 6-33 chars, `^[a-z][a-z0-9-]{4,31}[a-z0-9]$` |
| `clusters` | `object[]` | One or more cluster configurations. | Minimum 1 item |
| `clusters[].clusterId` | `string` | Unique cluster identifier within the instance. | 6-30 chars, `^[a-z][a-z0-9-]{4,28}[a-z0-9]$` |
| `clusters[].zone` | `string` | Zone where the cluster is deployed. Each cluster must be in a different zone. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | Instance name | Human-readable display name for the instance. |
| `deletionProtection` | `bool` | `true` | Prevents accidental destruction. Set to `false` before destroying. |
| `forceDestroy` | `bool` | `false` | Delete all backups when destroying the instance. |
| `clusters[].numNodes` | `int` | auto | Fixed number of nodes. Mutually exclusive with `autoscalingConfig`. |
| `clusters[].storageType` | `string` | `SSD` | Storage type: `SSD` (low latency) or `HDD` (lower cost, batch workloads). Immutable. |
| `clusters[].kmsKeyName` | `string` | — | Cloud KMS key for CMEK encryption. Can reference a GcpKmsKey via `valueFrom`. Immutable. |
| `clusters[].nodeScalingFactor` | `string` | `NodeScalingFactor1X` | Node scaling granularity: `NodeScalingFactor1X` or `NodeScalingFactor2X`. Immutable. |
| `clusters[].autoscalingConfig.minNodes` | `int` | — | Minimum nodes for autoscaling. Required when autoscaling is configured. `>= 1` |
| `clusters[].autoscalingConfig.maxNodes` | `int` | — | Maximum nodes for autoscaling. Must be `>= minNodes`. |
| `clusters[].autoscalingConfig.cpuTarget` | `int` | — | Target CPU utilization percentage. Range: 10-80. |
| `clusters[].autoscalingConfig.storageTarget` | `int` | — | Target storage per node in GB. SSD: 2560-5120, HDD: 8192-16384. |

## Examples

### Single Cluster with Fixed Nodes

A Bigtable instance with a fixed 3-node SSD cluster for predictable workloads:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpBigtableInstance
metadata:
  name: analytics-bt
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpBigtableInstance.analytics-bt
spec:
  projectId:
    value: my-gcp-project
  instanceName: analytics-bigtable
  clusters:
    - clusterId: analytics-cluster
      zone: us-central1-a
      numNodes: 3
```

### Autoscaling Cluster

A Bigtable instance that scales between 2 and 20 nodes based on CPU utilization:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpBigtableInstance
metadata:
  name: timeseries-bt
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpBigtableInstance.timeseries-bt
spec:
  projectId:
    value: my-gcp-project
  instanceName: timeseries-bigtable
  displayName: Time Series Production
  deletionProtection: true
  clusters:
    - clusterId: timeseries-us-c1a
      zone: us-central1-a
      autoscalingConfig:
        minNodes: 2
        maxNodes: 20
        cpuTarget: 65
```

### Multi-Cluster Replication

Two clusters in different zones for automatic replication and failover:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpBigtableInstance
metadata:
  name: ha-bigtable
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpBigtableInstance.ha-bigtable
spec:
  projectId:
    value: my-gcp-project
  instanceName: ha-bigtable-prod
  displayName: HA Production Bigtable
  deletionProtection: true
  clusters:
    - clusterId: ha-cluster-zone-a
      zone: us-central1-a
      autoscalingConfig:
        minNodes: 3
        maxNodes: 30
        cpuTarget: 60
    - clusterId: ha-cluster-zone-b
      zone: us-central1-b
      autoscalingConfig:
        minNodes: 3
        maxNodes: 30
        cpuTarget: 60
```

### CMEK Encrypted with Foreign Key Reference

Clusters encrypted with a Cloud KMS key, referenced from a GcpKmsKey resource:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpBigtableInstance
metadata:
  name: encrypted-bt
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpBigtableInstance.encrypted-bt
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: prod-project
      field: status.outputs.project_id
  instanceName: encrypted-bigtable
  deletionProtection: true
  clusters:
    - clusterId: encrypted-cluster-a
      zone: us-central1-a
      kmsKeyName:
        valueFrom:
          kind: GcpKmsKey
          name: bigtable-cmek-key
          field: status.outputs.key_id
      autoscalingConfig:
        minNodes: 3
        maxNodes: 30
        cpuTarget: 65
        storageTarget: 4096

```

### Full-Featured Production

All optional fields configured for an enterprise deployment:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpBigtableInstance
metadata:
  name: enterprise-bt
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: data-platform
    pulumi.openmcf.org/stack.name: prod.GcpBigtableInstance.enterprise-bt
spec:
  projectId:
    value: acme-data-prod
  instanceName: enterprise-bigtable
  displayName: Enterprise Data Platform
  deletionProtection: true
  forceDestroy: false
  clusters:
    - clusterId: enterprise-us-c1a
      zone: us-central1-a
      storageType: SSD
      nodeScalingFactor: NodeScalingFactor1X
      kmsKeyName: projects/acme-data-prod/locations/us-central1/keyRings/bigtable-kr/cryptoKeys/bigtable-key
      autoscalingConfig:
        minNodes: 5
        maxNodes: 50
        cpuTarget: 60
        storageTarget: 4096
    - clusterId: enterprise-us-c1b
      zone: us-central1-b
      storageType: SSD
      nodeScalingFactor: NodeScalingFactor1X
      kmsKeyName: projects/acme-data-prod/locations/us-central1/keyRings/bigtable-kr/cryptoKeys/bigtable-key
      autoscalingConfig:
        minNodes: 5
        maxNodes: 50
        cpuTarget: 60
        storageTarget: 4096
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `instance_id` | `string` | Fully qualified instance resource name. Format: `projects/{project}/instances/{instance}` |
| `instance_name` | `string` | Short instance name, same as the `instanceName` spec input. Used by Bigtable client libraries along with the project ID to connect. |

## Related Components

- [GcpProject](/docs/catalog/gcp/project) — project where the instance is created
- [GcpKmsKey](/docs/catalog/gcp/kms-key) — encryption key for CMEK-protected clusters
- [GcpKmsKeyRing](/docs/catalog/gcp/kms-key-ring) — key ring containing the CMEK key
- [GcpVpc](/docs/catalog/gcp/vpc) — network infrastructure for Private Service Connect (if applicable)
