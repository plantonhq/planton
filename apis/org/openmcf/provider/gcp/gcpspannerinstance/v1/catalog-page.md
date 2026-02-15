# GCP Spanner Instance

Deploys a Google Cloud Spanner instance with configurable compute capacity via fixed nodes, processing units, or autoscaling. Supports PROVISIONED and FREE_INSTANCE types, three editions (STANDARD, ENTERPRISE, ENTERPRISE_PLUS), and automatic backup scheduling for databases created within the instance.

## What Gets Created

When you deploy a GcpSpannerInstance resource, OpenMCF provisions:

- **Spanner Instance** — a `google_spanner_instance` resource with the specified instance configuration, compute capacity, and edition
- **Framework Labels** — OpenMCF metadata labels applied to the instance (resource kind, name, organization, environment)

## Prerequisites

- **GCP credentials** configured via environment variables or OpenMCF provider config
- **A GCP project** where the Spanner instance will be created
- **Spanner API enabled** in the target project (`spanner.googleapis.com`)
- **Billing account** attached to the project (required even for FREE_INSTANCE — limited to one per billing account)

## Quick Start

Create a file `spanner-instance.yaml`:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpSpannerInstance
metadata:
  name: my-spanner
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpSpannerInstance.my-spanner
spec:
  projectId:
    value: my-gcp-project-123
  instanceName: my-spanner
  config: regional-us-central1
  displayName: My Spanner Instance
  numNodes: 1
```

Deploy:

```shell
openmcf apply -f spanner-instance.yaml
```

This creates a single-node Spanner instance in the `us-central1` region.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `StringValueOrRef` | GCP project ID. Can reference a GcpProject resource via `valueFrom`. | Required |
| `instanceName` | `string` | Unique name for the instance. Immutable after creation. | 6-30 chars, pattern `^[a-z][-a-z0-9]*[a-z0-9]$` |
| `config` | `string` | Instance configuration defining replication topology (e.g., `regional-us-central1`, `nam-eur-asia1`). Immutable after creation. | Required |
| `displayName` | `string` | Human-readable display name. Must be unique within the project. | 4-30 chars |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `numNodes` | `int32` | — | Number of nodes. Each node = ~10,000 QPS reads, 10 TB storage. Mutually exclusive with `processingUnits` and `autoscalingConfig`. |
| `processingUnits` | `int32` | — | Processing units for finer-grained sizing. 1 node = 1000 PUs. Mutually exclusive with `numNodes` and `autoscalingConfig`. |
| `autoscalingConfig` | `object` | — | Autoscaling configuration with min/max bounds and utilization targets. Mutually exclusive with `numNodes` and `processingUnits`. |
| `autoscalingConfig.autoscalingLimits.minNodes` | `int32` | — | Minimum nodes. Use with `maxNodes`. |
| `autoscalingConfig.autoscalingLimits.maxNodes` | `int32` | — | Maximum nodes. Must be >= `minNodes`. |
| `autoscalingConfig.autoscalingLimits.minProcessingUnits` | `int32` | — | Minimum PUs. Use with `maxProcessingUnits`. |
| `autoscalingConfig.autoscalingLimits.maxProcessingUnits` | `int32` | — | Maximum PUs. Must be >= `minProcessingUnits`. |
| `autoscalingConfig.autoscalingTargets.highPriorityCpuUtilizationPercent` | `int32` | — | CPU target (0-100). Recommended: 65. |
| `autoscalingConfig.autoscalingTargets.storageUtilizationPercent` | `int32` | — | Storage target (0-100). Recommended: 80. |
| `instanceType` | `string` | `PROVISIONED` | `PROVISIONED` or `FREE_INSTANCE`. Free instances cannot set capacity, edition, or AUTOMATIC backups. |
| `edition` | `string` | — | `STANDARD`, `ENTERPRISE`, or `ENTERPRISE_PLUS`. Cannot be set for FREE_INSTANCE. |
| `defaultBackupScheduleType` | `string` | `NONE` | `NONE` or `AUTOMATIC`. Controls automatic backup creation for new databases. Cannot be `AUTOMATIC` for FREE_INSTANCE. |
| `forceDestroy` | `bool` | `false` | When `true`, deletes all backups when destroying the instance. Required if backups exist. |

## Examples

### Free Instance for Development

Zero-cost instance for development and testing:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpSpannerInstance
metadata:
  name: dev-spanner
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpSpannerInstance.dev-spanner
spec:
  projectId:
    value: my-gcp-project-123
  instanceName: dev-spanner
  config: regional-us-central1
  displayName: Dev Spanner
  instanceType: FREE_INSTANCE
```

### Regional Production with Autoscaling

Production instance that scales automatically between 1 and 5 nodes:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpSpannerInstance
metadata:
  name: prod-spanner
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpSpannerInstance.prod-spanner
spec:
  projectId:
    value: my-gcp-project-123
  instanceName: prod-spanner
  config: regional-us-central1
  displayName: Production Spanner
  edition: ENTERPRISE
  defaultBackupScheduleType: AUTOMATIC
  autoscalingConfig:
    autoscalingLimits:
      minNodes: 1
      maxNodes: 5
    autoscalingTargets:
      highPriorityCpuUtilizationPercent: 65
      storageUtilizationPercent: 80
```

### Multi-Region Enterprise Plus

Globally distributed instance with 99.999% availability SLA:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpSpannerInstance
metadata:
  name: global-spanner
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpSpannerInstance.global-spanner
spec:
  projectId:
    value: my-gcp-project-123
  instanceName: global-spanner
  config: nam-eur-asia1
  displayName: Global Spanner
  numNodes: 3
  edition: ENTERPRISE_PLUS
  defaultBackupScheduleType: AUTOMATIC
```

### Using Foreign Key References

Reference a GcpProject managed by OpenMCF instead of hardcoding the project ID:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpSpannerInstance
metadata:
  name: ref-spanner
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpSpannerInstance.ref-spanner
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      fieldPath: status.outputs.project_id
  instanceName: ref-spanner
  config: regional-us-central1
  displayName: Referenced Spanner
  numNodes: 1
  edition: ENTERPRISE
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `instance_id` | `string` | Fully qualified instance ID (`projects/{project}/instances/{name}`) |
| `instance_name` | `string` | Short instance name, used by GcpSpannerDatabase to reference this instance |
| `state` | `string` | Instance state: `CREATING` or `READY` |

## Related Components

- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the project where the Spanner instance is created
- [GcpSpannerDatabase](/docs/catalog/gcp/gcpspannerdatabase) — creates databases within this instance (references `instance_name` output)
