# GcpSpannerInstance

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpSpannerInstanceSpec defines the configuration for a Google Cloud
Spanner instance.

Cloud Spanner is Google's fully managed, globally distributed, strongly
consistent relational database. The INSTANCE is the unit of compute and
storage allocation: it pins a geographic topology (the instance
configuration) and a capacity envelope that all databases on it share.
Databases (GcpSpannerDatabase) and backup schedules
(GcpSpannerBackupSchedule) are separate composable resources that
reference this instance by name.

Capacity is expressed exactly one way:
  - num_nodes — coarse units (1 node = 1000 processing units)
  - processing_units — fine-grained (multiples of 100 below 1000)
  - autoscaling_config — Spanner manages capacity within bounds

Important behavioral notes:

  - instance_name, config, and project are IMMUTABLE — changing any of
    them recreates the instance (and everything on it). Everything else,
    including capacity, edition, and autoscaling, updates in place.

  - Capacity changes (nodes/processing units, or switching to/from
    autoscaling) are online operations with no downtime.

  - Edition upgrades (STANDARD → ENTERPRISE → ENTERPRISE_PLUS) apply in
    place. Downgrades require first disabling features of the higher
    edition.

  - FREE_INSTANCE provides one zero-cost instance per billing account
    with limited capacity (about 10 GB storage). It must not set capacity
    fields, edition, or an AUTOMATIC default backup schedule, and it can
    be upgraded to PROVISIONED in place (but never back).

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSpannerInstance
metadata:
  name: orders-spanner
spec:
  # project_id omitted: falls back to the provider's default project.
  config: regional-us-central1
  displayName: Orders Spanner
  processingUnits: 100
  edition: STANDARD
  # Explicit DELETE (the provider default) proves the round-trip; use
  # PREVENT for instances a whole topology depends on.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.instanceName` | `string` | yes |  |  |
| `spec.config` | `string` | yes |  |  |
| `spec.displayName` | `string` | yes |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.numNodes` | `int32` |  |  |  |
| `spec.processingUnits` | `int32` |  |  |  |
| `spec.autoscalingConfig` | `GcpSpannerInstanceAutoscalingConfig` |  |  |  |
| `spec.autoscalingConfig.autoscalingLimits` | `GcpSpannerInstanceAutoscalingLimits` | yes |  |  |
| `spec.autoscalingConfig.autoscalingLimits.minNodes` | `int32` |  |  |  |
| `spec.autoscalingConfig.autoscalingLimits.maxNodes` | `int32` |  |  |  |
| `spec.autoscalingConfig.autoscalingLimits.minProcessingUnits` | `int32` |  |  |  |
| `spec.autoscalingConfig.autoscalingLimits.maxProcessingUnits` | `int32` |  |  |  |
| `spec.autoscalingConfig.autoscalingTargets` | `GcpSpannerInstanceAutoscalingTargets` |  |  |  |
| `spec.autoscalingConfig.autoscalingTargets.highPriorityCpuUtilizationPercent` | `int32` |  |  |  |
| `spec.autoscalingConfig.autoscalingTargets.storageUtilizationPercent` | `int32` |  |  |  |
| `spec.autoscalingConfig.autoscalingTargets.totalCpuUtilizationPercent` | `int32` |  |  |  |
| `spec.autoscalingConfig.asymmetricAutoscalingOptions` | `[]GcpSpannerInstanceAsymmetricAutoscalingOption` |  |  |  |
| `spec.autoscalingConfig.asymmetricAutoscalingOptions[].replicaLocation` | `string` | yes |  |  |
| `spec.autoscalingConfig.asymmetricAutoscalingOptions[].overrides` | `GcpSpannerInstanceAsymmetricAutoscalingOverrides` | yes |  |  |
| `spec.autoscalingConfig.asymmetricAutoscalingOptions[].overrides.minNodes` | `int32` |  |  |  |
| `spec.autoscalingConfig.asymmetricAutoscalingOptions[].overrides.maxNodes` | `int32` |  |  |  |
| `spec.autoscalingConfig.asymmetricAutoscalingOptions[].overrides.minProcessingUnits` | `int32` |  |  |  |
| `spec.autoscalingConfig.asymmetricAutoscalingOptions[].overrides.maxProcessingUnits` | `int32` |  |  |  |
| `spec.autoscalingConfig.asymmetricAutoscalingOptions[].overrides.autoscalingTargetHighPriorityCpuUtilizationPercent` | `int32` |  |  |  |
| `spec.autoscalingConfig.asymmetricAutoscalingOptions[].overrides.autoscalingTargetTotalCpuUtilizationPercent` | `int32` |  |  |  |
| `spec.autoscalingConfig.asymmetricAutoscalingOptions[].overrides.disableHighPriorityCpuAutoscaling` | `bool` |  |  |  |
| `spec.autoscalingConfig.asymmetricAutoscalingOptions[].overrides.disableTotalCpuAutoscaling` | `bool` |  |  |  |
| `spec.instanceType` | `string` |  |  |  |
| `spec.edition` | `string` |  |  |  |
| `spec.defaultBackupScheduleType` | `string` |  |  |  |
| `spec.forceDestroy` | `bool` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the Spanner instance is created in. Accepts a literal
project ID or a reference to a GcpProject resource. If omitted, the
provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.instanceName

`string` · required

Unique identifier of the Spanner instance in GCP. Immutable. If not
specified, defaults to metadata.name. Must be 6-30 characters: start
with a lowercase letter, contain only lowercase letters, digits, and
hyphens, and end with a letter or digit. This is the value downstream
databases and backup schedules reference.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"6","maxLen":"30","pattern":"^[a-z][-a-z0-9]*[a-z0-9]$"}}

### spec.config

`string` · required

Instance configuration defining geographic placement and replication
topology. Immutable. Regional configs (e.g. "regional-us-central1")
give the lowest latency in one region; multi-region configs (e.g.
"nam6", "nam-eur-asia1") replicate across regions for higher
availability and enable the 99.999% SLA on ENTERPRISE_PLUS. This is
the single most consequential choice on the instance — it cannot be
changed without recreating the instance and moving the data.

- rule: {"required":true}

### spec.displayName

`string` · required

Human-readable display name shown in the GCP console.
Must be 4-30 characters.

- rule: {"required":true,"string":{"minLen":"4","maxLen":"30"}}

### spec.labels

`map<string, string>`

Labels applied to the instance for cost attribution and organization.
Merged with Planton's platform labels (which win on key conflicts).

### spec.numNodes

`int32`

Number of nodes allocated to the instance. Each node provides roughly
10,000 QPS of reads or 2,000 QPS of writes and 10 TB of storage.
Mutable — capacity changes apply online. Mutually exclusive with
processing_units and autoscaling_config; must not be set for
FREE_INSTANCE. If no capacity field is set for a PROVISIONED instance,
GCP defaults to 1 node.

### spec.processingUnits

`int32`

Number of processing units allocated to the instance — the
fine-grained alternative to nodes (1 node = 1000 processing units).
Values below 1000 must be multiples of 100 (the smallest billable
Spanner footprint is 100). Mutable. Mutually exclusive with num_nodes
and autoscaling_config; must not be set for FREE_INSTANCE.

### spec.autoscalingConfig

`GcpSpannerInstanceAutoscalingConfig`

Managed autoscaling: Spanner adjusts capacity within bounds based on
CPU and storage utilization, including per-replica overrides for
multi-region instances. Mutually exclusive with num_nodes and
processing_units; must not be set for FREE_INSTANCE.

### spec.autoscalingConfig.autoscalingLimits

`GcpSpannerInstanceAutoscalingLimits` · required

Required. The floor and ceiling the autoscaler operates within.

- rule: {"required":true}
- rule: use either nodes (min_nodes + max_nodes) or processing_units (min_processing_units + max_processing_units), not both
- rule: max_nodes must be >= min_nodes and both must be positive
- rule: max_processing_units must be >= min_processing_units and both must be positive

### spec.autoscalingConfig.autoscalingLimits.minNodes

`int32`

Minimum number of nodes the autoscaler may scale down to. Use together
with max_nodes. Each node is 1000 processing units and provides roughly
10,000 QPS of reads or 2,000 QPS of writes plus 10 TB of storage.

### spec.autoscalingConfig.autoscalingLimits.maxNodes

`int32`

Maximum number of nodes the autoscaler may scale up to.
Must be >= min_nodes.

### spec.autoscalingConfig.autoscalingLimits.minProcessingUnits

`int32`

Minimum number of processing units the autoscaler may scale down to.
Use together with max_processing_units. Values below 1000 must be
multiples of 100; values 1000 and above must be multiples of 1000.

### spec.autoscalingConfig.autoscalingLimits.maxProcessingUnits

`int32`

Maximum number of processing units the autoscaler may scale up to.
Must be >= min_processing_units.

### spec.autoscalingConfig.autoscalingTargets

`GcpSpannerInstanceAutoscalingTargets`

Utilization targets that trigger scaling decisions.
If not set, GCP uses default targets.

### spec.autoscalingConfig.autoscalingTargets.highPriorityCpuUtilizationPercent

`int32`

Target percentage of high-priority CPU utilization (user reads/writes,
as opposed to background maintenance work). When the instance runs
hotter than this, the autoscaler adds capacity. Google recommends 65
for regional configurations and lower (around 45) for multi-region
configurations, where replication headroom matters during failover.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.autoscalingConfig.autoscalingTargets.storageUtilizationPercent

`int32`

Target percentage of storage utilization. When storage crosses this
threshold, the autoscaler adds capacity regardless of CPU (Spanner
couples storage capacity to compute capacity). Google recommends 80
to leave headroom for growth spikes.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.autoscalingConfig.autoscalingTargets.totalCpuUtilizationPercent

`int32`

Target percentage of TOTAL CPU utilization — user traffic plus
Spanner's background maintenance work — as a second CPU signal
alongside high_priority_cpu_utilization_percent. Useful when
background work (change streams, backups) is a meaningful share of
the load and scaling on user traffic alone would run the instance
hot. Scale 0 (no utilization) to 100 (full utilization).

- rule: {"int32":{"lte":100,"gte":0}}

### spec.autoscalingConfig.asymmetricAutoscalingOptions

`[]GcpSpannerInstanceAsymmetricAutoscalingOption`

Per-replica-location overrides for multi-region instances: scale a
read-heavy region independently instead of sizing every region for the
hottest one. Each entry selects one replica location and overrides its
capacity bounds, CPU targets, or both.

### spec.autoscalingConfig.asymmetricAutoscalingOptions[].replicaLocation

`string` · required

The replica location (a region of the multi-region configuration,
e.g. "europe-west1") whose autoscaling this option overrides.

- rule: {"required":true}

### spec.autoscalingConfig.asymmetricAutoscalingOptions[].overrides

`GcpSpannerInstanceAsymmetricAutoscalingOverrides` · required

The per-replica tuning applied at this location: capacity bounds
(nodes or processing units) replacing the instance-wide
autoscaling_limits, replica-local CPU targets, or switches disabling
a CPU signal for this replica.

- rule: {"required":true}
- rule: use either nodes (min_nodes + max_nodes) or processing units (min_processing_units + max_processing_units), not both
- rule: max_nodes must be >= min_nodes and both must be positive
- rule: max_processing_units must be >= min_processing_units and both must be positive

### spec.autoscalingConfig.asymmetricAutoscalingOptions[].overrides.minNodes

`int32`

Minimum number of nodes for the selected replica location.
Use together with max_nodes; mutually exclusive with the
processing-unit bounds.

### spec.autoscalingConfig.asymmetricAutoscalingOptions[].overrides.maxNodes

`int32`

Maximum number of nodes for the selected replica location.
Must be >= min_nodes.

### spec.autoscalingConfig.asymmetricAutoscalingOptions[].overrides.minProcessingUnits

`int32`

Minimum number of processing units for the selected replica
location — the fine-grained alternative to node bounds. Per-replica
PU bounds must be multiples of 1000 (unlike instance-wide limits,
which allow multiples of 100 below 1000). Use together with
max_processing_units; mutually exclusive with the node bounds.

### spec.autoscalingConfig.asymmetricAutoscalingOptions[].overrides.maxProcessingUnits

`int32`

Maximum number of processing units for the selected replica
location. Must be >= min_processing_units and a multiple of 1000.

### spec.autoscalingConfig.asymmetricAutoscalingOptions[].overrides.autoscalingTargetHighPriorityCpuUtilizationPercent

`int32`

Target high-priority CPU utilization percentage for THIS replica,
overriding the instance-wide autoscaling_targets value there.
Scale 0-100.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.autoscalingConfig.asymmetricAutoscalingOptions[].overrides.autoscalingTargetTotalCpuUtilizationPercent

`int32`

Target total CPU utilization percentage for THIS replica, overriding
the instance-wide autoscaling_targets value there. Scale 0-100.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.autoscalingConfig.asymmetricAutoscalingOptions[].overrides.disableHighPriorityCpuAutoscaling

`bool`

Disable high-priority CPU autoscaling for this replica: the
instance-wide high_priority_cpu_utilization_percent target is
ignored there and only the remaining signals drive scaling.

### spec.autoscalingConfig.asymmetricAutoscalingOptions[].overrides.disableTotalCpuAutoscaling

`bool`

Disable total CPU autoscaling for this replica: the instance-wide
total_cpu_utilization_percent target is ignored there.

### spec.instanceType

`string`

Instance type. PROVISIONED (default) requires explicit capacity via
one of the capacity fields. FREE_INSTANCE provisions the billing
account's one zero-cost development instance and must not set any
capacity field. Upgrading FREE_INSTANCE → PROVISIONED works in place;
the reverse does not exist.

- rule: instance_type must be PROVISIONED or FREE_INSTANCE

### spec.edition

`string`

Edition controlling available features and SLA level. Mutable —
upgrades apply in place. STANDARD: cost-optimized, single-region
feature set. ENTERPRISE: adds granular sizing, asymmetric autoscaling,
and incremental backups. ENTERPRISE_PLUS: 99.999% multi-region SLA and
advanced compliance features. Cannot be set for FREE_INSTANCE.

- rule: edition must be STANDARD, ENTERPRISE, or ENTERPRISE_PLUS

### spec.defaultBackupScheduleType

`string`

Default backup schedule type for NEW databases created on this
instance. NONE (default): new databases get no automatic backup
schedule. AUTOMATIC: GCP attaches a default backup schedule to each
new database (explicit GcpSpannerBackupSchedule resources give full
control instead). Cannot be AUTOMATIC for FREE_INSTANCE.

- rule: default_backup_schedule_type must be NONE or AUTOMATIC

### spec.forceDestroy

`bool`

Whether destroying the instance also deletes all backups held on it.
When false (default), destroy fails if any database on the instance
has backups — a safety net against losing the last restore point. Set
true only when the backups are intentionally disposable.

### spec.deletionPolicy

`string`

Deletion policy for the instance — what happens when this resource
is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the instance is deleted, taking every database and
               backup on it with it (backups additionally gated by
               force_destroy above)
  "PREVENT" -- destroy FAILS; protects the instance every database
               in the topology depends on
  "ABANDON" -- the instance is removed from management but left
               running (and billing) in GCP with its data intact

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `capacity_mutual_exclusion`: only one of num_nodes, processing_units, or autoscaling_config may be set
- `free_instance_no_capacity`: FREE_INSTANCE must not set num_nodes, processing_units, or autoscaling_config
- `free_instance_no_edition`: edition cannot be set when instance_type is FREE_INSTANCE
- `free_instance_no_automatic_backup`: default_backup_schedule_type cannot be AUTOMATIC when instance_type is FREE_INSTANCE

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpSpannerInstance, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.instance_id` | `string` | Fully qualified instance ID. Format: projects/{project}/instances/{instance_name} This is the canonical identifier used for IAM bindings and API calls. |
| `status.outputs.instance_name` | `string` | Short instance name. This is the value that GcpSpannerDatabase, GcpSpannerBackupSchedule, and other downstream resources use to reference this instance. |
| `status.outputs.state` | `string` | Instance state: CREATING or READY. CREATING indicates the instance is being provisioned. READY indicates the instance is available for use. |
| `status.outputs.config` | `string` | The instance configuration the instance was created with (e.g. "regional-us-central1", "nam6") — the geographic topology handle auditors and capacity planners ask for. Always the plain config name: the Spanner API reads the value back as the fully qualified projects/{project}/instanceConfigs/{name} path, and both engines normalize to the last path segment. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpSpannerBackupSchedule | `spec.instance` | `status.outputs.instance_name` |
| GcpSpannerDatabase | `spec.instance` | `status.outputs.instance_name` |

## See Also

- [Overview](../README.md)
