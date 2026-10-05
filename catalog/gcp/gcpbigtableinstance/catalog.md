# GCP Bigtable Instance

Deploys a Cloud Bigtable instance with one or more clusters, configurable scaling (fixed nodes, autoscaling, or auto-allocated), SSD or HDD storage, optional CMEK encryption per cluster, and multi-zone replication for high availability. The instance bundles its clusters because an instance without at least one cluster cannot store or serve data; tables are a separate kind with their own lifecycle.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Bigtable Admin API enablement** (`bigtableadmin.googleapis.com`) on the target project (never disabled on destroy)
- **Bigtable Instance** -- a managed instance in the specified GCP project, serving as the logical container for data with deletion protection and optional display name
- **Bigtable Clusters** -- one or more clusters within the instance, each placed in a specific zone with independent scaling configuration (fixed node count, autoscaling with CPU/storage targets, or auto-allocated)
- **CMEK Encryption** -- created only when `kmsKeyName` is set on a cluster; encrypts data at rest with a customer-managed Cloud KMS key (immutable per cluster)
- **GCP Labels** -- resource metadata labels (resource name, kind, organization, environment) applied at the instance level for tracking and governance

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the Bigtable instance will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **Cloud KMS key** (if using CMEK) -- the key region must match the cluster zone's region. The Bigtable service account must have `roles/cloudkms.cryptoKeyEncrypterDecrypter` on the key.

## Deploy

### Console

Open the deployment store, find **GCP Bigtable Instance**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Dev Single Node** preset in the [Presets](#presets) tab to pre-populate a minimal development configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBigtableInstance
metadata:
  name: events-store
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  instanceName: events-store-prod
  clusters:
    - clusterId: events-cluster-a
      zone: us-central1-a
      numNodes: 3
```

```shell
planton apply -f bigtable-instance.yaml
```

This creates a Bigtable instance with a single 3-node SSD cluster, Google-managed encryption, and deletion protection enabled (default). No autoscaling or multi-zone replication is configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the instance to a GCP project deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
```

The InfraPipeline resolves the dependency graph, deploys the project first, then provisions the Bigtable instance with the resolved project ID.

## Key Configuration

These are the most important decisions when configuring a Bigtable instance. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Cluster scaling mode** -- Each cluster supports three scaling modes. Fixed (`numNodes`) provides predictable capacity. Autoscaling (`autoscalingConfig`) adjusts nodes between `minNodes` and `maxNodes` based on `cpuTarget` (10-80%) and optional `storageTarget`. Auto-allocated (neither set) lets Bigtable manage nodes based on data footprint. Autoscaling and fixed nodes are mutually exclusive per cluster.

**Storage type** -- `storageType` defaults to `SSD` (lowest latency, recommended for most workloads). Use `HDD` for large batch-analytics workloads where latency is less critical and cost optimization is a priority. Immutable after cluster creation.

**Multi-cluster replication** -- Add multiple clusters in different zones for automatic replication and failover. Bigtable client libraries handle routing transparently. Each cluster must be in a unique zone. Up to 8 clusters across regions are supported.

**Deletion protection** -- `deletionProtection` defaults to `true`, preventing accidental instance destruction. Set to `false` only for development instances that need easy teardown. Enable `forceDestroy` if the instance has backups that should be deleted during destroy.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpKmsKey** (optional, per cluster) | `clusters[].kmsKeyName` | `status.outputs.key_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `instance_id` | Fully qualified instance resource name (`projects/{p}/instances/{i}`) | Monitoring dashboards, IAM bindings |
| `instance_name` | Short instance name for Bigtable client library connections | Application connection configuration |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Development single node** -- Single auto-allocated cluster with SSD storage and deletion protection disabled. Minimal cost for development, CI/CD, and prototyping. Start from the **Dev Single Node** preset.

**HA production** -- Two clusters in different zones with autoscaling (3-30 nodes, 65% CPU target). Automatic replication and failover for production workloads with variable traffic. Start from the **HA Production** preset.

**Enterprise encrypted** -- Two CMEK-encrypted clusters with aggressive autoscaling (5-50 nodes, 60% CPU, explicit storage targets). Designed for regulated industries requiring customer-managed encryption and large-scale capacity. Start from the **Enterprise Encrypted** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the Bigtable instance is created
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) -- provides the Cloud KMS key for per-cluster CMEK encryption
- [**GCP Bigtable Table**](/infra-catalog/gcp-bigtable-table) -- tables created on this instance, referencing its `instance_name` output