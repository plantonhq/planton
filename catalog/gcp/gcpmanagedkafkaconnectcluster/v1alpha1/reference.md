# GcpManagedKafkaConnectCluster

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpManagedKafkaConnectClusterSpec defines a Managed Service for Apache
Kafka Connect cluster (`google_managed_kafka_connect_cluster`) -- Google-
operated Kafka Connect workers attached to one Kafka cluster, running the
connectors that move data between Kafka and other systems (Pub/Sub,
BigQuery, Cloud Storage, another Kafka cluster through MirrorMaker 2).

The connectors themselves are GcpManagedKafkaConnector blocks, owned by
the teams whose data they move.

Immutable: project_id, location, connect_cluster_id (a change replaces
the Connect cluster and every connector on it). The attached Kafka
cluster, capacity, networks, and labels update in place.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpManagedKafkaConnectCluster
metadata:
  name: events-connect
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  kafkaCluster:
    value: projects/my-gcp-project/locations/us-central1/clusters/events
  capacityConfig:
    vcpuCount: 6
    memoryBytes: 25769803776
  networkConfigs:
    - primarySubnet:
        value: projects/my-gcp-project/regions/us-central1/subnetworks/kafka
      # Lets a MirrorMaker 2 connector resolve a second Kafka cluster.
      dnsDomainNames:
        - legacy-events.us-central1.managedkafka.my-gcp-project.cloud.goog
  labels:
    team: data-platform
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.connectClusterId` | `string` |  |  |  |
| `spec.kafkaCluster` | `string \| valueFrom` | yes |  | GcpManagedKafkaCluster (`status.outputs.name`) |
| `spec.capacityConfig` | `GcpManagedKafkaConnectClusterCapacity` | yes |  |  |
| `spec.capacityConfig.vcpuCount` | `int64` |  |  |  |
| `spec.capacityConfig.memoryBytes` | `int64` |  |  |  |
| `spec.networkConfigs` | `[]GcpManagedKafkaConnectClusterNetworkConfig` | yes |  |  |
| `spec.networkConfigs[].primarySubnet` | `string \| valueFrom` | yes |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.networkConfigs[].dnsDomainNames` | `[]string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the Connect cluster lives in: a literal project ID or
a GcpProject reference. If omitted, the provider's default project is
used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The region the workers run in, e.g. "us-central1". Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.connectClusterId

`string`

The Connect cluster's ID. Defaults to metadata.name. Immutable.

### spec.kafkaCluster

`string | valueFrom` · required

The Kafka cluster the workers attach to (their internal topics and the
connectors' default cluster): a GcpManagedKafkaCluster reference (its
name output) or a literal full path
projects/{project}/locations/{location}/clusters/{cluster}. Mutable in
place.

- references: GcpManagedKafkaCluster (`status.outputs.name`)
- rule: a literal kafka_cluster must be the full path projects/{project}/locations/{location}/clusters/{cluster}
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpManagedKafkaCluster, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.capacityConfig

`GcpManagedKafkaConnectClusterCapacity` · required

vCPUs and memory for the workers.

- rule: {"required":true}
- rule: memory_bytes must keep a vCPU:GiB ratio between 1:1 and 1:8 (vcpu_count x 1073741824 to vcpu_count x 8589934592)

### spec.capacityConfig.vcpuCount

`int64`

vCPUs for the Connect workers -- at least 3. Billed per vCPU-hour.
Sent as a decimal string.

- rule: {"int64":{"gte":"3"}}

### spec.capacityConfig.memoryBytes

`int64`

Memory for the Connect workers, in BYTES as a plain number -- at least
3 GiB (3221225472), at a vCPU:GiB ratio between 1:1 and 1:8. Billed per
GiB-hour. Sent as a decimal string.

- rule: {"int64":{"gte":"3221225472"}}

### spec.networkConfigs

`[]GcpManagedKafkaConnectClusterNetworkConfig` · required

The VPC networks the workers reach -- at least one, at most 10.

- rule: {"repeated":{"minItems":"1","maxItems":"10"}}

### spec.networkConfigs[].primarySubnet

`string | valueFrom` · required

The subnet the workers' interface is placed in: a GcpSubnetwork
reference (its self link, which the modules trim to the
projects/{project}/regions/{region}/subnetworks/{subnet} form Google
requires) or a literal in either form. It must be in the Connect
cluster's region, and its CIDR range must be private (RFC 1918).

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.networkConfigs[].dnsDomainNames

`[]string`

Extra DNS domains from this network the workers may resolve. For
MirrorMaker 2, add the target Kafka cluster's DNS domain -- the
bootstrap address without its leading "bootstrap." label and its port
(e.g. my-cluster.us-central1.managedkafka.my-project.cloud.goog).

### spec.labels

`map<string, string>`

Labels on the Connect cluster. The platform attribution labels are
added on top and win on a key conflict.

### spec.deletionPolicy

`string`

What happens to the Connect cluster when this resource is destroyed:
  "" / "DELETE" -- deleted, with its connectors
  "PREVENT"     -- destroy fails
  "ABANDON"     -- it leaves management and keeps running (and
                   billing) in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpManagedKafkaConnectCluster, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name -- projects/{project}/locations/{location}/connectClusters/{connect_cluster_id}. What connectors reference. |
| `status.outputs.connect_cluster_id` | `string` | The Connect cluster's ID (the last segment of name). |
| `status.outputs.location` | `string` | The region the workers run in. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.kafkaCluster` | GcpManagedKafkaCluster | `status.outputs.name` |
| `spec.networkConfigs[].primarySubnet` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpManagedKafkaConnector | `spec.connectCluster` | `status.outputs.name` |

## See Also

- [Overview](../README.md)
