# GcpRedisClusterEndpointSet

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpRedisClusterEndpointSetSpec registers the Private Service Connect
connections a consumer built by hand on a Memorystore for Redis Cluster
(`google_redis_cluster_user_created_connections`) -- Google's
"user-created connections", the way a cluster is reached from VPCs or
projects its own connectivity automation cannot place endpoints in.

The flow is three blocks in order: the cluster (usually without
psc_configs, so it only publishes service attachments), then in each
consumer VPC one reserved GcpAddress and one regional
GcpGlobalForwardingRule per service attachment (empty scheme, the
attachment handle as target), then this set naming every rule. Google
replaces the cluster's whole user-created endpoint list with this set on
every apply, so declare exactly one set per cluster and list every
consumer network in it. Removing a connection here stops the matching
forwarding rule from working; delete the rule in the same change.

Only the project is immutable; the set updates in place.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpRedisClusterEndpointSet
metadata:
  name: orders-cache-endpoints
spec:
  projectId:
    value: my-gcp-project
  # The cluster, created without pscConfigs so it publishes attachments.
  cluster:
    valueFrom:
      kind: GcpRedisCluster
      name: orders-cache
      fieldPath: status.outputs.name
  region: us-central1
  endpoints:
    # One consumer VPC: a connection for the discovery attachment and one
    # for the primary (a cluster with replicas adds a reader connection).
    - connections:
        - forwardingRule:
            valueFrom:
              kind: GcpGlobalForwardingRule
              name: orders-cache-disc
              fieldPath: status.outputs.self_link
          pscConnectionId:
            valueFrom:
              kind: GcpGlobalForwardingRule
              name: orders-cache-disc
              fieldPath: status.outputs.psc_connection_id
          address:
            valueFrom:
              kind: GcpAddress
              name: orders-cache-disc-ip
              fieldPath: status.outputs.address
          network:
            valueFrom:
              kind: GcpVpcNetwork
              name: consumer-vpc
              fieldPath: status.outputs.network_id
          serviceAttachment:
            valueFrom:
              kind: GcpRedisCluster
              name: orders-cache
              fieldPath: status.outputs.discovery_service_attachment
        - forwardingRule:
            valueFrom:
              kind: GcpGlobalForwardingRule
              name: orders-cache-prim
              fieldPath: status.outputs.self_link
          pscConnectionId:
            valueFrom:
              kind: GcpGlobalForwardingRule
              name: orders-cache-prim
              fieldPath: status.outputs.psc_connection_id
          address:
            valueFrom:
              kind: GcpAddress
              name: orders-cache-prim-ip
              fieldPath: status.outputs.address
          network:
            valueFrom:
              kind: GcpVpcNetwork
              name: consumer-vpc
              fieldPath: status.outputs.network_id
          serviceAttachment:
            valueFrom:
              kind: GcpRedisCluster
              name: orders-cache
              fieldPath: status.outputs.primary_service_attachment
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.cluster` | `string \| valueFrom` | yes |  | GcpRedisCluster (`status.outputs.name`) |
| `spec.region` | `string` | yes |  |  |
| `spec.endpoints` | `[]GcpRedisClusterEndpointSetEndpoint` | yes |  |  |
| `spec.endpoints[].connections` | `[]GcpRedisClusterEndpointSetConnection` | yes |  |  |
| `spec.endpoints[].connections[].forwardingRule` | `string \| valueFrom` | yes |  | GcpGlobalForwardingRule (`status.outputs.self_link`) |
| `spec.endpoints[].connections[].pscConnectionId` | `string \| valueFrom` | yes |  | GcpGlobalForwardingRule (`status.outputs.psc_connection_id`) |
| `spec.endpoints[].connections[].address` | `string \| valueFrom` | yes |  | GcpAddress (`status.outputs.address`) |
| `spec.endpoints[].connections[].network` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_id`) |
| `spec.endpoints[].connections[].serviceAttachment` | `string \| valueFrom` | yes |  | GcpRedisCluster (`status.outputs.discovery_service_attachment`) |
| `spec.endpoints[].connections[].projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the cluster lives in: a literal project ID or a
GcpProject reference. If omitted, the provider's default project is
used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.cluster

`string | valueFrom` · required

The cluster the connections belong to. A GcpRedisCluster reference
resolves to its full resource path (name output); a literal takes
either the full path or the bare cluster name. The modules derive the
bare name Google's resource expects.

- references: GcpRedisCluster (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpRedisCluster, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.region

`string` · required

The cluster's region (e.g. "us-central1").

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.endpoints

`[]GcpRedisClusterEndpointSetEndpoint` · required

One entry per consumer VPC, each carrying a connection per cluster
service attachment. This list IS the cluster's user-created endpoint
set: anything not listed is removed on apply.

- rule: {"repeated":{"minItems":"1"}}

### spec.endpoints[].connections

`[]GcpRedisClusterEndpointSetConnection` · required

The connections that make up this endpoint, one per cluster service
attachment, all in one consumer VPC.

- rule: {"repeated":{"minItems":"1"}}

### spec.endpoints[].connections[].forwardingRule

`string | valueFrom` · required

The consumer-side forwarding rule, as its URI. A GcpGlobalForwardingRule
reference (the regional PSC form: empty scheme, the cluster's service
attachment as target) resolves to its self_link output.

- references: GcpGlobalForwardingRule (`status.outputs.self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGlobalForwardingRule, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.endpoints[].connections[].pscConnectionId

`string | valueFrom` · required

The PSC connection ID Google assigned to that forwarding rule. The SAME
GcpGlobalForwardingRule as forwarding_rule, referenced on its
psc_connection_id output -- a reference names one output path, so the
rule is named twice.

- references: GcpGlobalForwardingRule (`status.outputs.psc_connection_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGlobalForwardingRule, name: <that resource's name>, fieldPath: status.outputs.psc_connection_id}} -- a bare string does not parse

### spec.endpoints[].connections[].address

`string | valueFrom` · required

The IP address the forwarding rule serves on the consumer network. A
GcpAddress reference (the reserved internal address the rule was given)
resolves to its address output.

- references: GcpAddress (`status.outputs.address`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpAddress, name: <that resource's name>, fieldPath: status.outputs.address}} -- a bare string does not parse

### spec.endpoints[].connections[].network

`string | valueFrom` · required

The consumer VPC network the address lives in, as
projects/{project}/global/networks/{name}. A GcpVpcNetwork reference
resolves to its network_id output.

- references: GcpVpcNetwork (`status.outputs.network_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_id}} -- a bare string does not parse

### spec.endpoints[].connections[].serviceAttachment

`string | valueFrom` · required

The cluster service attachment this connection targets, as
projects/{project}/regions/{region}/serviceAttachments/{id}. A
GcpRedisCluster reference resolves to its discovery_service_attachment
output by default; the connection for the primary or reader endpoint
names that handle through an explicit fieldPath
(status.outputs.primary_service_attachment,
status.outputs.reader_service_attachment).

- references: GcpRedisCluster (`status.outputs.discovery_service_attachment`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpRedisCluster, name: <that resource's name>, fieldPath: status.outputs.discovery_service_attachment}} -- a bare string does not parse

### spec.endpoints[].connections[].projectId

`string | valueFrom`

The consumer project the forwarding rule was created in. If omitted,
Google records the project it finds on the rule -- set it only when
the rule lives in a different project than the cluster.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.deletionPolicy

`string`

What happens to the registration when this resource is destroyed:
  "" / "DELETE" -- the connections are deregistered (the forwarding
                   rules themselves belong to their own blocks)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the registration leaves management but stays on
                   the cluster

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpRedisClusterEndpointSet, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.cluster_name` | `string` | Bare name of the cluster the connections were registered on -- the segment Google's resource is keyed by. |
| `status.outputs.endpoint_count` | `int32` | Number of consumer-network endpoints registered (one per VPC). |
| `status.outputs.connection_count` | `int32` | Number of individual PSC connections registered across every endpoint. |
| `status.outputs.region` | `string` | The cluster's region -- with cluster_name, what rebuilds the cluster's full resource path. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.cluster` | GcpRedisCluster | `status.outputs.name` |
| `spec.endpoints[].connections[].forwardingRule` | GcpGlobalForwardingRule | `status.outputs.self_link` |
| `spec.endpoints[].connections[].pscConnectionId` | GcpGlobalForwardingRule | `status.outputs.psc_connection_id` |
| `spec.endpoints[].connections[].address` | GcpAddress | `status.outputs.address` |
| `spec.endpoints[].connections[].network` | GcpVpcNetwork | `status.outputs.network_id` |
| `spec.endpoints[].connections[].serviceAttachment` | GcpRedisCluster | `status.outputs.discovery_service_attachment` |
| `spec.endpoints[].connections[].projectId` | GcpProject | `status.outputs.project_id` |

## See Also

- [Overview](../README.md)
