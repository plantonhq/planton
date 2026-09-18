# GcpAlloydbInstance

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpAlloydbInstanceSpec defines an AlloyDB instance (`google_alloydb_instance`)
attached to an existing cluster.

Read pool instances scale read traffic independently of the bundled primary
in GcpAlloydbCluster. PRIMARY and SECONDARY types exist for advanced
topologies; most presets target READ_POOL.

## Example

```yaml
# Exercises a READ_POOL instance offline: cluster by literal path, two read
# nodes, and TLS-only client connections.
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpAlloydbInstance
metadata:
  name: hack-orders-read-pool
spec:
  cluster:
    value: projects/my-project/locations/us-central1/clusters/hack-orders
  instanceId: orders-read-pool
  instanceType: READ_POOL
  cpuCount: 2
  # availabilityType stays empty on read pools — derived from nodeCount.
  readPoolConfig:
    nodeCount: 2
  requireConnectors: true
  sslMode: ENCRYPTED_ONLY
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.cluster` | `string \| valueFrom` | yes |  | GcpAlloydbCluster (`status.outputs.cluster_id`) |
| `spec.instanceId` | `string` | yes |  |  |
| `spec.instanceType` | `string` |  | `READ_POOL` |  |
| `spec.cpuCount` | `int32` |  |  |  |
| `spec.machineType` | `string` |  |  |  |
| `spec.readPoolConfig` | `GcpAlloydbInstanceReadPoolConfig` |  |  |  |
| `spec.readPoolConfig.nodeCount` | `int32` |  |  |  |
| `spec.availabilityType` | `string` |  |  |  |
| `spec.databaseFlags` | `map<string, string>` |  |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.queryInsightsConfig` | `GcpAlloydbInstanceQueryInsightsConfig` |  |  |  |
| `spec.queryInsightsConfig.queryPlansPerMinute` | `int32` |  |  |  |
| `spec.queryInsightsConfig.queryStringLength` | `int32` |  |  |  |
| `spec.queryInsightsConfig.recordApplicationTags` | `bool` |  |  |  |
| `spec.queryInsightsConfig.recordClientAddress` | `bool` |  |  |  |
| `spec.requireConnectors` | `bool` |  |  |  |
| `spec.sslMode` | `string` |  |  |  |
| `spec.activationPolicy` | `string` |  |  |  |
| `spec.enablePublicIp` | `bool` |  |  |  |
| `spec.enableOutboundPublicIp` | `bool` |  |  |  |
| `spec.authorizedExternalNetworks` | `[]GcpAlloydbInstanceAuthorizedExternalNetwork` |  |  |  |
| `spec.authorizedExternalNetworks[].cidrRange` | `string` | yes |  |  |
| `spec.pscInstanceConfig` | `GcpAlloydbInstancePscInstanceConfig` |  |  |  |
| `spec.pscInstanceConfig.allowedConsumerProjects` | `[]string` |  |  |  |
| `spec.pscInstanceConfig.pscAutoConnections` | `[]GcpAlloydbInstancePscAutoConnection` |  |  |  |
| `spec.pscInstanceConfig.pscAutoConnections[].consumerNetwork` | `string` |  |  |  |
| `spec.pscInstanceConfig.pscAutoConnections[].consumerProject` | `string` |  |  |  |
| `spec.pscInstanceConfig.pscInterfaceConfigs` | `[]GcpAlloydbInstancePscInterfaceConfig` |  |  |  |
| `spec.pscInstanceConfig.pscInterfaceConfigs[].networkAttachmentResource` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.gceZone` | `string` |  |  |  |
| `spec.connectionPoolConfig` | `GcpAlloydbInstanceConnectionPoolConfig` |  |  |  |
| `spec.connectionPoolConfig.enabled` | `bool` |  |  |  |
| `spec.connectionPoolConfig.flags` | `map<string, string>` |  |  |  |
| `spec.allocatedIpRangeOverride` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the AlloyDB cluster.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.cluster

`string | valueFrom` · required

The AlloyDB cluster this instance belongs to. Accepts the full cluster
resource path or a reference to a GcpAlloydbCluster resource. Immutable.

- references: GcpAlloydbCluster (`status.outputs.cluster_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpAlloydbCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_id}} -- a bare string does not parse

### spec.instanceId

`string` · required

The instance ID within the cluster. Immutable.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"63","pattern":"^[a-z][a-z0-9-]{0,61}[a-z0-9]$"}}

### spec.instanceType

`string` · optional (explicit presence)

Instance role: PRIMARY, READ_POOL, or SECONDARY. Immutable.
Presets default to READ_POOL for read scaling.

- default: `READ_POOL`
- rule: instance_type must be PRIMARY, READ_POOL, or SECONDARY

### spec.cpuCount

`int32`

Number of CPUs. GCP selects the machine family. Mutually exclusive with
machine_type.

### spec.machineType

`string`

Explicit machine type (e.g. "n2-highmem-4"). Mutually exclusive with
cpu_count.

### spec.readPoolConfig

`GcpAlloydbInstanceReadPoolConfig`

Read pool sizing. Required when instance_type is READ_POOL.

### spec.readPoolConfig.nodeCount

`int32`

Read capacity — number of nodes in the read pool instance.

- rule: {"int32":{"gte":1}}

### spec.availabilityType

`string`

ZONAL or REGIONAL placement for PRIMARY/SECONDARY instances. GCP
defaults to REGIONAL when unset. Must stay empty on READ_POOL
instances: read-pool availability is DERIVED from node_count (1 node =
ZONAL, 2+ nodes = REGIONAL spread across zones) and the AlloyDB API
does not store a sent value — the stored object omits the field, so
any explicit value produces a perpetual re-plan diff (live-verified
against a single-node read pool).

- rule: availability_type must be ZONAL, REGIONAL, or AVAILABILITY_TYPE_UNSPECIFIED

### spec.databaseFlags

`map<string, string>`

PostgreSQL database flags as key-value pairs.

### spec.displayName

`string`

Human-readable display name.

### spec.queryInsightsConfig

`GcpAlloydbInstanceQueryInsightsConfig`

Query insights configuration.

### spec.queryInsightsConfig.queryPlansPerMinute

`int32`

Number of query execution plans captured per minute. Range: 0-20.

- rule: {"int32":{"lte":20,"gte":0}}

### spec.queryInsightsConfig.queryStringLength

`int32`

Maximum length of the query string stored in insights. Range: 256-4500.
0 means unset — GCP applies its default (1024).

- rule: query_string_length must be between 256 and 4500 (or 0 for the GCP default)

### spec.queryInsightsConfig.recordApplicationTags

`bool`

Whether to record application tags for queries.

### spec.queryInsightsConfig.recordClientAddress

`bool`

Whether to record the client IP address for each query.

### spec.requireConnectors

`bool`

When true, only AlloyDB Auth Proxy / Language Connectors may connect.

### spec.sslMode

`string`

SSL mode: ENCRYPTED_ONLY or ALLOW_UNENCRYPTED_AND_ENCRYPTED.

- rule: ssl_mode must be ENCRYPTED_ONLY or ALLOW_UNENCRYPTED_AND_ENCRYPTED

### spec.activationPolicy

`string`

Instance activation: ALWAYS keeps the instance running (the default
posture); NEVER stops it. Flipping ALWAYS→NEVER→ALWAYS is the
stop/start lever — a stopped instance keeps its configuration and
storage but serves nothing and stops billing for compute. Mind the
ordering restrictions (stop read pools before the primary).

- rule: activation_policy must be ALWAYS, NEVER, or ACTIVATION_POLICY_UNSPECIFIED

### spec.enablePublicIp

`bool`

Enable a public IP on the instance.

### spec.enableOutboundPublicIp

`bool`

Enable outbound public IP for the instance.

### spec.authorizedExternalNetworks

`[]GcpAlloydbInstanceAuthorizedExternalNetwork`

CIDR ranges allowed to reach the public IP. Requires enable_public_ip.

### spec.authorizedExternalNetworks[].cidrRange

`string` · required

- rule: {"string":{"minLen":"1"}}

### spec.pscInstanceConfig

`GcpAlloydbInstancePscInstanceConfig`

Private Service Connect configuration.

### spec.pscInstanceConfig.allowedConsumerProjects

`[]string`

Consumer project numbers allowed to create PSC endpoints.

### spec.pscInstanceConfig.pscAutoConnections

`[]GcpAlloydbInstancePscAutoConnection`

PSC service automation connections.

### spec.pscInstanceConfig.pscAutoConnections[].consumerNetwork

`string`

Consumer network, e.g. "projects/vpc-host/global/networks/default".

### spec.pscInstanceConfig.pscAutoConnections[].consumerProject

`string`

Consumer project ID (not project number).

### spec.pscInstanceConfig.pscInterfaceConfigs

`[]GcpAlloydbInstancePscInterfaceConfig`

PSC interfaces for outbound connectivity (0 or 1 supported by AlloyDB).

### spec.pscInstanceConfig.pscInterfaceConfigs[].networkAttachmentResource

`string`

Network attachment resource in the consumer project.

### spec.labels

`map<string, string>`

User-defined labels on the instance (cost attribution, team ownership,
environment tagging). Merged with the platform's attribution labels;
on key conflicts the platform labels win. Mutable in place.

### spec.annotations

`map<string, string>`

Unstructured metadata stored on the instance (annotations, not labels —
not used for billing filtering). Mutable in place.

### spec.gceZone

`string`

Pin a ZONAL instance to a specific Compute Engine zone (e.g.
"us-central1-a"). Only valid when availability_type is ZONAL — GCP
rejects it on REGIONAL instances; leave empty to let GCP pick a zone
with available capacity. Mutable: changing it live-migrates the
instance to the new zone.

### spec.connectionPoolConfig

`GcpAlloydbInstanceConnectionPoolConfig`

AlloyDB managed connection pooling (built-in pooler). Mutable in place.

### spec.connectionPoolConfig.enabled

`bool`

Turn managed connection pooling on or off. Mutable in place.

### spec.connectionPoolConfig.flags

`map<string, string>`

Pooler flags, keyed by flag name WITHOUT the "connection-pooling-"
prefix and with underscores instead of dashes (GCP's documented
convention for this provider surface): e.g. the flag
"connection-pooling-pool-mode" is set as key "pool_mode". Only
applied while enabled is true.

### spec.allocatedIpRangeOverride

`string`

Draw this instance's private IPs from a specific Private Service
Access allocated range (RFC 1035 name, e.g.
"google-managed-services-default") instead of the range the parent
cluster uses. Immutable: changing it destroys and recreates the
instance.

- rule: allocated_ip_range_override must be an RFC 1035 range name (1-63 chars: [a-z]([-a-z0-9]*[a-z0-9])?)

### spec.deletionPolicy

`string`

What happens to the instance in GCP when this resource is destroyed.
  "DELETE"  -- (GCP's default when unset) the instance is deleted;
               the parent cluster and its data survive
  "PREVENT" -- destroy FAILS; protects serving capacity applications
               still connect to
  "ABANDON" -- the instance is removed from management but keeps
               running (and billing) in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `machine_config_mutual_exclusion`: only one of cpu_count or machine_type may be set
- `read_pool_requires_node_count`: READ_POOL instances require read_pool_config.node_count >= 1
- `read_pool_config_only_for_read_pool`: read_pool_config applies to READ_POOL instances only
- `authorized_networks_require_public_ip`: authorized_external_networks requires enable_public_ip
- `availability_type_not_for_read_pool`: availability_type applies to PRIMARY/SECONDARY instances only — read-pool availability is derived from node_count (1 node = ZONAL, 2+ nodes = REGIONAL) and the API does not store a sent value
- `gce_zone_requires_zonal`: gce_zone can only be set on ZONAL instances — GCP rejects it when availability_type is REGIONAL (the default)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpAlloydbInstance, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.instance_name` | `string` | Fully qualified instance resource name. Format: projects/{project}/locations/{location}/clusters/{cluster}/instances/{instance} |
| `status.outputs.ip_address` | `string` | Private IP address of the instance (primary connection endpoint). |
| `status.outputs.state` | `string` | Current state of the instance (e.g. READY, CREATING). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.cluster` | GcpAlloydbCluster | `status.outputs.cluster_id` |

## See Also

- [Overview](../README.md)
