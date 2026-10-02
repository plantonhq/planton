# GcpManagedKafkaCluster

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpManagedKafkaClusterSpec defines a Managed Service for Apache Kafka
cluster (`google_managed_kafka_cluster`) -- Google-operated Apache Kafka
brokers in one region, reachable from the VPC networks you attach.
Clients authenticate with Google IAM (SASL OAUTHBEARER, port 9092) or,
with tls_config, mutual TLS (port 9192).

What lives elsewhere, as their own blocks:
  - topics: GcpManagedKafkaTopic (owned by the teams that produce to them)
  - access rules: GcpManagedKafkaAcl (one per resource pattern)
  - Kafka Connect: GcpManagedKafkaConnectCluster and
    GcpManagedKafkaConnector

The bootstrap address clients connect to is fixed for the cluster's
lifetime but its format can differ between clusters, and the pinned
Pulumi SDK cannot read it yet, so it is not an output: read it once with
`gcloud managed-kafka clusters describe CLUSTER --location=LOCATION
--format="value(bootstrapAddress)"`. Internet (public) access is not yet
offered here for the same SDK reason; clients reach the cluster from an
attached VPC.

Immutable: project_id, location, cluster_id, kms_key (a change replaces
the cluster and its data). Capacity, networks, rebalancing, TLS, and
labels update in place.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpManagedKafkaCluster
metadata:
  name: events
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  clusterId: events
  # 3 vCPU and 12 GiB across the cluster (4 GiB per vCPU).
  capacityConfig:
    vcpuCount: 3
    memoryBytes: 12884901888
  brokerDiskSizeGib: 200
  networkConfigs:
    - subnet:
        value: projects/my-gcp-project/regions/us-central1/subnetworks/kafka
  rebalanceMode: AUTO_REBALANCE_ON_SCALE_UP
  labels:
    team: data-platform
  deletionPolicy: PREVENT
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.clusterId` | `string` |  |  |  |
| `spec.capacityConfig` | `GcpManagedKafkaClusterCapacity` | yes |  |  |
| `spec.capacityConfig.vcpuCount` | `int64` |  |  |  |
| `spec.capacityConfig.memoryBytes` | `int64` |  |  |  |
| `spec.brokerDiskSizeGib` | `int64` |  |  |  |
| `spec.networkConfigs` | `[]GcpManagedKafkaClusterNetworkConfig` | yes |  |  |
| `spec.networkConfigs[].subnet` | `string \| valueFrom` | yes |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.kmsKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.rebalanceMode` | `string` |  |  |  |
| `spec.tlsConfig` | `GcpManagedKafkaClusterTlsConfig` |  |  |  |
| `spec.tlsConfig.sslPrincipalMappingRules` | `string` |  |  |  |
| `spec.tlsConfig.caPools` | `[]string \| valueFrom` |  |  | GcpPrivateCaPool (`status.outputs.name`) |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the cluster lives in: a literal project ID or a
GcpProject reference. If omitted, the provider's default project is
used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The region the brokers run in, e.g. "us-central1". Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.clusterId

`string`

The cluster's ID -- 1-63 characters, RFC 1035 (lowercase letters,
digits, hyphens; starts with a letter). Defaults to metadata.name.
Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?$"}}

### spec.capacityConfig

`GcpManagedKafkaClusterCapacity` · required

vCPUs and memory for the whole cluster.

- rule: {"required":true}
- rule: memory_bytes must be between 1 GiB and 8 GiB per vCPU (vcpu_count x 1073741824 to vcpu_count x 8589934592)

### spec.capacityConfig.vcpuCount

`int64`

vCPUs provisioned across the cluster -- at least 3. Billed per
vCPU-hour whether or not traffic flows. Sent as a decimal string.
Mutable in place (scaling up can trigger a rebalance, see
rebalance_mode).

- rule: {"int64":{"gte":"3"}}

### spec.capacityConfig.memoryBytes

`int64`

Memory provisioned across the cluster, in BYTES as a plain number
(3 GiB = 3221225472) -- between 1 GiB and 8 GiB per vCPU. Billed per
GiB-hour. Sent as a decimal string. Mutable in place.

- rule: {"int64":{"gte":"1073741824"}}

### spec.brokerDiskSizeGib

`int64`

Disk per broker in GiB -- at least 100. Empty leaves Google's default.
Sent as a decimal string.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int64":{"gte":"100"}}

### spec.networkConfigs

`[]GcpManagedKafkaClusterNetworkConfig` · required

The VPC networks the cluster is reachable from -- at least one, at
most 10, one subnet per network.

- rule: {"repeated":{"minItems":"1","maxItems":"10"}}

### spec.networkConfigs[].subnet

`string | valueFrom` · required

The subnet the cluster is reachable from: a GcpSubnetwork reference
(its self link, which the modules trim to the
projects/{project}/regions/{region}/subnetworks/{subnet} form Google
requires) or a literal in either form. It must be in the cluster's
region; the project may differ (a Shared VPC host). Only one subnet
per VPC network.

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.kmsKey

`string | valueFrom`

A Cloud KMS key encrypting the cluster's data at rest (CMEK): a
GcpKmsKey reference or a literal
projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}.
The key must be in the cluster's region, and the Managed Kafka service
agent needs roles/cloudkms.cryptoKeyEncrypterDecrypter on it. Empty
uses Google-managed keys. Immutable.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.rebalanceMode

`string`

Whether Google moves partitions onto new brokers when the cluster
scales up:
  NO_REBALANCE               -- never (Google's default when empty)
  AUTO_REBALANCE_ON_SCALE_UP -- rebalance automatically after a
                                scale-up, spreading load onto the new
                                brokers

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["NO_REBALANCE","AUTO_REBALANCE_ON_SCALE_UP"]}}

### spec.tlsConfig

`GcpManagedKafkaClusterTlsConfig`

Mutual TLS for clients. See the message comment for how omitting,
declaring, and emptying it differ.

### spec.tlsConfig.sslPrincipalMappingRules

`string`

How the Distinguished Name of a client certificate becomes the short
principal name ACLs match (Kafka's ssl.principal.mapping.rules broker
setting, same syntax). Empty keeps Kafka's default behavior. Changing
it triggers a rolling restart of the brokers.

### spec.tlsConfig.caPools

`[]string | valueFrom`

Certificate Authority Service CA pools whose certificates the brokers
trust for client authentication -- GcpPrivateCaPool references (their
full names) or literals projects/{project}/locations/{location}/caPools/{pool},
in any project or location. At most 10. Setting at least one enables
mTLS.

- references: GcpPrivateCaPool (`status.outputs.name`)
- rule: a literal CA pool must be projects/{project}/locations/{location}/caPools/{pool}
- rule: {"repeated":{"maxItems":"10"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpPrivateCaPool, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.labels

`map<string, string>`

Labels on the cluster. The platform attribution labels are added on
top and win on a key conflict.

### spec.deletionPolicy

`string`

What happens to the cluster when this resource is destroyed:
  "" / "DELETE" -- deleted, with every topic and message in it
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the cluster leaves management and keeps running
                   (and billing) in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpManagedKafkaCluster, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name -- projects/{project}/locations/{location}/clusters/{cluster_id}. What topics, ACLs, and Kafka Connect clusters reference. |
| `status.outputs.cluster_id` | `string` | The cluster's ID (the last segment of name). |
| `status.outputs.location` | `string` | The region the cluster runs in. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.networkConfigs[].subnet` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |
| `spec.kmsKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.tlsConfig.caPools` | GcpPrivateCaPool | `status.outputs.name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpManagedKafkaAcl | `spec.cluster` | `status.outputs.name` |
| GcpManagedKafkaConnectCluster | `spec.kafkaCluster` | `status.outputs.name` |
| GcpManagedKafkaTopic | `spec.cluster` | `status.outputs.name` |

## See Also

- [Overview](../README.md)
