# AwsMemcachedElasticache

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsMemcachedElasticacheSpec defines the desired configuration for an AWS
ElastiCache cluster running the Memcached engine.

This component provisions a fully managed, distributed in-memory cache using
the Memcached protocol. Memcached is optimized for simple key-value caching
with sub-millisecond latency and horizontal scaling across multiple nodes.

Key differences from Redis/Valkey (AwsRedisElasticache):

- **No replication** — Memcached distributes keys across nodes via consistent
  hashing. Each key lives on exactly one node. If a node fails, its keys are
  lost and must be re-populated from the data source.

- **No persistence** — there are no snapshots, no AOF, no RDB. Memcached is
  a pure volatile cache.

- **No authentication** — Memcached has no AUTH mechanism. Security relies
  entirely on VPC network isolation (security groups and private subnets).

- **No encryption at rest** — only in-transit encryption is available, and
  only on engine version 1.6.12 or later.

- **Horizontal scaling** — add or remove nodes (1–40) to increase or decrease
  cache capacity. Node additions are non-disruptive; node removals may evict
  keys that hashed to the removed node.

- **Node type changes force recreation** — Memcached does not support
  vertical scaling in-place. Changing `node_type` destroys and recreates the
  entire cluster, losing all cached data.

Use Memcached when you need a simple, high-throughput distributed cache and
do not require data persistence, replication, or authentication.

Notes:
- `port` is ForceNew. Default: 11211.
- Credentials, region, and deployment workflow live outside this spec in
  IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsMemcachedElasticache
metadata:
  name: test-memcached
  org: test-org
  env: dev
  id: test-memcached-dev
spec:
  region: us-west-2
  engineVersion: "1.6.22"
  nodeType: cache.t3.micro
  numCacheNodes: 1
  # Pin all nodes to one AZ (mutually exclusive with the per-node
  # preferred_availability_zones list and with az_mode "cross-az").
  availabilityZone: us-west-2a
  # Presence-typed: unset lets AWS's enabled-by-default stand; an explicit
  # value pins it.
  autoMinorVersionUpgrade: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.engineVersion` | `string` |  |  |  |
| `spec.nodeType` | `string` | yes |  |  |
| `spec.numCacheNodes` | `int32` |  |  |  |
| `spec.azMode` | `string` |  |  |  |
| `spec.port` | `int32` |  | `11211` |  |
| `spec.transitEncryptionEnabled` | `bool` |  |  |  |
| `spec.subnetIds` | `[]string \| valueFrom` |  |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.subnetGroupName` | `string` |  |  |  |
| `spec.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.networkType` | `string` |  |  |  |
| `spec.ipDiscovery` | `string` |  |  |  |
| `spec.parameterGroupFamily` | `string` |  |  |  |
| `spec.parameters` | `[]AwsMemcachedElasticacheParameter` |  |  |  |
| `spec.parameters[].name` | `string` | yes |  |  |
| `spec.parameters[].value` | `string` | yes |  |  |
| `spec.parameterGroupName` | `string` |  |  |  |
| `spec.maintenanceWindow` | `string` |  |  |  |
| `spec.applyImmediately` | `bool` |  |  |  |
| `spec.autoMinorVersionUpgrade` | `bool` |  | `true` |  |
| `spec.notificationTopicArn` | `string \| valueFrom` |  |  | AwsSnsTopic (`status.outputs.topic_arn`) |
| `spec.preferredAvailabilityZones` | `[]string` |  |  |  |
| `spec.availabilityZone` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.engineVersion

`string`

Memcached engine version to deploy. Uses three-part versioning:
"1.6.22", "1.6.17", "1.5.16", etc. Leave empty to use the AWS default —
a versionless manifest never goes stale.
Transit encryption requires version 1.6.12 or later.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^\\d+\\.\\d+\\.\\d+$"}}

### spec.nodeType

`string` · required

ElastiCache node type. Determines CPU, memory, and network capacity.
Examples: "cache.t3.micro" (dev), "cache.r7g.large" (production).
Changing node_type forces cluster recreation — Memcached does not support
vertical scaling in-place.

- rule: {"required":true}

### spec.numCacheNodes

`int32`

Number of cache nodes in the cluster. Memcached distributes keys across
all nodes via consistent hashing. Range: 1–40. Default: 1.

- rule: {"int32":{"lte":40,"gte":1}}

### spec.azMode

`string`

AZ distribution mode. "single-az" places all nodes in one AZ (default).
"cross-az" distributes nodes across multiple AZs for resilience.
cross-az requires num_cache_nodes > 1.

### spec.port

`int32` · optional (explicit presence)

Port on which the cluster accepts connections. Default: 11211.
This is a ForceNew attribute — changing it destroys and recreates the
cluster.

- default: `11211`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.transitEncryptionEnabled

`bool`

Enable encryption in transit (TLS) for all client connections.
Requires Memcached engine version 1.6.12 or later. Attempting to enable
this on earlier versions will result in an AWS API error.
Note: Memcached does NOT support encryption at rest.

### spec.subnetIds

`[]string | valueFrom`

Subnet IDs for the ElastiCache subnet group. Provide subnets in at least
two AZs when using cross-az mode. A subnet group is created automatically
from these subnets. Mutually exclusive with `subnet_group_name`.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.subnetGroupName

`string`

Name of an EXISTING ElastiCache subnet group to place the cluster in,
instead of building one from `subnet_ids`. Bring-your-own for
organizations that manage subnet groups centrally. ForceNew — changing
the subnet group replaces the cluster.

### spec.securityGroupIds

`[]string | valueFrom`

VPC security groups to attach to the cluster nodes. Controls network-level
access to the Memcached endpoint. Since Memcached has no authentication,
security groups are the primary access control mechanism.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.networkType

`string`

IP addressing for the cluster's network. Values: "ipv4" (default),
"ipv6", "dual_stack". ForceNew — changing the network type replaces the
cluster. Dual-stack requires subnets with both IPv4 and IPv6 CIDRs.

### spec.ipDiscovery

`string`

Which address family DNS discovery returns to clients. Values: "ipv4",
"ipv6". Only meaningful alongside a dual-stack `network_type`; updates
in place, letting clients migrate address families without replacing
the cluster.

### spec.parameterGroupFamily

`string`

Parameter group family for custom parameters. Required when `parameters`
is provided. Examples: "memcached1.6", "memcached1.5", "memcached1.4".

### spec.parameters

`[]AwsMemcachedElasticacheParameter`

Custom cache parameters to apply via a managed parameter group. Common
Memcached parameters include: chunk_size, chunk_size_growth_factor,
max_simultaneous_connections, binding_protocol. Mutually exclusive with
`parameter_group_name`.

### spec.parameters[].name

`string` · required

Parameter name (e.g., "chunk_size", "binding_protocol").

- rule: {"required":true}

### spec.parameters[].value

`string` · required

Parameter value (e.g., "96", "auto").

- rule: {"required":true}

### spec.parameterGroupName

`string`

Name of an EXISTING parameter group to use instead of managing
parameters here. Bring-your-own for organizations that share one tuned
group across many caches. Mutually exclusive with `parameters`.

### spec.maintenanceWindow

`string`

Weekly maintenance window in UTC. Format: "ddd:hh24:mi-ddd:hh24:mi".
Example: "sun:05:00-sun:06:00". Leave empty for AWS-assigned default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(mon|tue|wed|thu|fri|sat|sun):([01][0-9]|2[0-3]):[0-5][0-9]-(mon|tue|wed|thu|fri|sat|sun):([01][0-9]|2[0-3]):[0-5][0-9]$"}}

### spec.applyImmediately

`bool`

Apply changes immediately instead of waiting for the next maintenance
window. May cause brief downtime for some operations.

### spec.autoMinorVersionUpgrade

`bool` · optional (explicit presence)

Automatically apply minor engine version upgrades during maintenance
windows. AWS enables this by default: leave unset to keep the default,
set false to pin the running minor version explicitly, set true to pin
the opt-in. Presence matters — unset is forwarded to AWS as "decide",
never as false.

- default: `true`

### spec.notificationTopicArn

`string | valueFrom`

SNS topic ARN for cluster event notifications (node additions, removals,
maintenance events, etc.).

- references: AwsSnsTopic (`status.outputs.topic_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSnsTopic, name: <that resource's name>, fieldPath: status.outputs.topic_arn}} -- a bare string does not parse

### spec.preferredAvailabilityZones

`[]string`

Preferred Availability Zones for the cache nodes. When provided, the list
length must match num_cache_nodes. Nodes are placed in the specified AZs
in order. Leave empty for AWS-managed AZ distribution. Mutually exclusive
with `availability_zone` (the single-AZ pin).

### spec.availabilityZone

`string`

Pin ALL cache nodes to one Availability Zone (e.g. "us-west-2a") —
AZ-local latency for a client fleet that lives in a single zone.
ForceNew — changing the pin replaces the cluster. Leave empty for an
AWS-chosen AZ. Mutually exclusive with `preferred_availability_zones`
(per-node placement) and with az_mode "cross-az" — a single-AZ pin IS
single-az mode.

## Validation Rules

- `az_mode_valid_values`: az_mode must be 'single-az' or 'cross-az' when set
- `cross_az_requires_multi_node`: az_mode 'cross-az' requires num_cache_nodes > 1
- `az_list_matches_node_count`: preferred_availability_zones length must match num_cache_nodes when provided
- `parameters_require_family`: parameter_group_family is required when parameters are provided
- `network_type_valid_values`: network_type must be 'ipv4', 'ipv6', or 'dual_stack' when set
- `ip_discovery_valid_values`: ip_discovery must be 'ipv4' or 'ipv6' when set
- `subnet_arms_mutual_exclusion`: subnet_ids and subnet_group_name are mutually exclusive — build a subnet group from subnets or bring an existing group
- `parameter_arms_mutual_exclusion`: parameters and parameter_group_name are mutually exclusive — manage parameters here or bring an existing group
- `az_placement_mutual_exclusion`: availability_zone and preferred_availability_zones are mutually exclusive placement controls — pin all nodes to one AZ or place nodes per-AZ
- `az_pin_forbids_cross_az`: availability_zone pins all nodes to one AZ — it cannot be combined with az_mode 'cross-az'; use preferred_availability_zones for cross-AZ placement

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsMemcachedElasticache, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.cluster_id` | `string` | The identifier of the ElastiCache cluster. |
| `status.outputs.cluster_address` | `string` | The DNS name of the Memcached auto-discovery endpoint (without port). Clients that support auto-discovery connect here to automatically discover all nodes in the cluster. Empty for single-node clusters. |
| `status.outputs.configuration_endpoint` | `string` | The full configuration endpoint in "address:port" format. Memcached clients use this for auto-discovery of cluster topology. This is the recommended connection endpoint for multi-node clusters. |
| `status.outputs.arn` | `string` | The Amazon Resource Name of the ElastiCache cluster. Used in IAM policies and cross-service permissions. |
| `status.outputs.port` | `int32` | The port on which the cluster accepts connections. |
| `status.outputs.subnet_group_name` | `string` | The name of the ElastiCache subnet group associated with this cluster. Only populated when `subnet_ids` were provided and a subnet group was created by the module. |
| `status.outputs.parameter_group_name` | `string` | The name of the custom parameter group associated with this cluster. Only populated when `parameters` were provided and a parameter group was created by the module. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.notificationTopicArn` | AwsSnsTopic | `status.outputs.topic_arn` |

## See Also

- [Overview](../README.md)
