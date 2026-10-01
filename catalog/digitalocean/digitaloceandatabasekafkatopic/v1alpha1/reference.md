# DigitalOceanDatabaseKafkaTopic

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

DigitalOceanDatabaseKafkaTopicSpec models the full
digitalocean_database_kafka_topic resource surface: a topic on a
DigitalOcean managed Kafka cluster, with the complete per-topic
configuration block.

Topic creation is asynchronous on DigitalOcean's side (the API may
acknowledge before the topic is queryable); partition count and
replication factor and every config leaf update in place. DigitalOcean
never reports the partition count back after async partition changes, so
imports and drift checks treat it as configuration-only.

## Example

```yaml
# Reference manifests for DigitalOceanDatabaseKafkaTopic --
# protovalidate-valid, embedded as the reference page's Example block, and
# the documents the offline tofu plans render. Two documents: a minimal
# topic on API defaults, and a tuned compacted topic exercising the config
# block's leaf classes.
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseKafkaTopic
metadata:
  name: orders-events
spec:
  # Literal cluster UUID; use valueFrom to reference a
  # DigitalOceanDatabaseCluster resource instead.
  cluster:
    value: aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee
  topicName: orders-events
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseKafkaTopic
metadata:
  name: customer-snapshots
spec:
  cluster:
    value: aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee
  topicName: customer-snapshots
  partitionCount: 12
  replicationFactor: 3
  config:
    cleanupPolicy: compact
    compressionType: zstd
    minInsyncReplicas: 2
    minCleanableDirtyRatio: 0.4
    maxCompactionLagMs: 604800000
    segmentBytes: 209715200
    messageTimestampType: log_append_time
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.cluster` | `string \| valueFrom` | yes |  | DigitalOceanDatabaseCluster (`status.outputs.cluster_id`) |
| `spec.topicName` | `string` | yes |  |  |
| `spec.partitionCount` | `int32` |  |  |  |
| `spec.replicationFactor` | `int32` |  |  |  |
| `spec.config` | `DigitalOceanDatabaseKafkaTopicConfig` |  |  |  |
| `spec.config.cleanupPolicy` | `string` |  |  |  |
| `spec.config.compressionType` | `string` |  |  |  |
| `spec.config.deleteRetentionMs` | `uint64` |  |  |  |
| `spec.config.fileDeleteDelayMs` | `uint64` |  |  |  |
| `spec.config.flushMessages` | `uint64` |  |  |  |
| `spec.config.flushMs` | `uint64` |  |  |  |
| `spec.config.indexIntervalBytes` | `uint64` |  |  |  |
| `spec.config.maxCompactionLagMs` | `uint64` |  |  |  |
| `spec.config.maxMessageBytes` | `uint64` |  |  |  |
| `spec.config.messageDownConversionEnable` | `bool` |  |  |  |
| `spec.config.messageFormatVersion` | `string` |  |  |  |
| `spec.config.messageTimestampDifferenceMaxMs` | `int64` |  |  |  |
| `spec.config.messageTimestampType` | `string` |  |  |  |
| `spec.config.minCleanableDirtyRatio` | `double` |  |  |  |
| `spec.config.minCompactionLagMs` | `uint64` |  |  |  |
| `spec.config.minInsyncReplicas` | `int32` |  |  |  |
| `spec.config.preallocate` | `bool` |  |  |  |
| `spec.config.retentionBytes` | `int64` |  |  |  |
| `spec.config.retentionMs` | `int64` |  |  |  |
| `spec.config.segmentBytes` | `uint64` |  |  |  |
| `spec.config.segmentIndexBytes` | `uint64` |  |  |  |
| `spec.config.segmentJitterMs` | `uint64` |  |  |  |
| `spec.config.segmentMs` | `uint64` |  |  |  |

## Field Details

### spec.cluster

`string | valueFrom` · required

The Kafka database cluster to create the topic in. Use a literal
cluster UUID or a reference to a DigitalOceanDatabaseCluster resource
(the cluster must run the kafka engine -- DigitalOcean rejects topic
calls on other engines). Any Kafka plan works, Basic included (unlike
the schema registry, which needs a General Purpose plan). Changing it
replaces the topic.

- references: DigitalOceanDatabaseCluster (`status.outputs.cluster_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanDatabaseCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_id}} -- a bare string does not parse

### spec.topicName

`string` · required

Name of the Kafka topic. Unique within the cluster; the name IS the
topic's API identity. Changing it replaces the topic and drops the old
topic's messages.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.partitionCount

`int32` · optional (explicit presence)

(Optional) Number of partitions for the topic, 3 to 2048. When unset,
DigitalOcean creates 3. Updates apply in place, but Kafka only ever
ADDS partitions -- lowering the count is rejected by the API. Partition
changes apply asynchronously and DigitalOcean does not report the live
count back, so a grown count is invisible to drift detection.

- rule: {"int32":{"lte":2048,"gte":3}}

### spec.replicationFactor

`int32` · optional (explicit presence)

(Optional) Replication factor for the topic, at least 2. The ceiling is
the cluster's node count (API-enforced). When unset, DigitalOcean
creates 2 replicas. Updates apply in place.

- rule: {"int32":{"gte":2}}

### spec.config

`DigitalOceanDatabaseKafkaTopicConfig`

(Optional) Per-topic Kafka configuration. Every leaf is optional; an
unset leaf defers to the Kafka server default, which DigitalOcean
reports back after create. NOTE: when this message is present, the
provider seeds cleanup_policy to "compact_delete" unless the leaf says
otherwise -- set cleanup_policy explicitly to avoid surprises.

### spec.config.cleanupPolicy

`string`

(Optional) Log cleanup policy. One of delete (discard old segments),
compact (keep the latest value per key), or compact_delete (both).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["delete","compact","compact_delete"]}}

### spec.config.compressionType

`string`

(Optional) Final compression type for the topic. "producer" retains
whatever compression the producer used; "uncompressed" disables
compression.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["snappy","gzip","lz4","producer","uncompressed","zstd"]}}

### spec.config.deleteRetentionMs

`uint64` · optional (explicit presence)

(Optional) How long (ms) delete tombstone markers are retained on
compacted topics.

### spec.config.fileDeleteDelayMs

`uint64` · optional (explicit presence)

(Optional) Delay (ms) before deleting a segment file from the
filesystem.

### spec.config.flushMessages

`uint64` · optional (explicit presence)

(Optional) Number of messages accumulated on a partition before an
fsync is forced.

### spec.config.flushMs

`uint64` · optional (explicit presence)

(Optional) Maximum time (ms) a message is kept in memory before an
fsync is forced.

### spec.config.indexIntervalBytes

`uint64` · optional (explicit presence)

(Optional) Interval (bytes) at which entries are added to the offset
index.

### spec.config.maxCompactionLagMs

`uint64` · optional (explicit presence)

(Optional) Maximum time (ms) a message remains ineligible for
compaction.

### spec.config.maxMessageBytes

`uint64` · optional (explicit presence)

(Optional) Largest record batch size (bytes) the topic accepts.

### spec.config.messageDownConversionEnable

`bool` · optional (explicit presence)

(Optional) Whether the broker may down-convert messages for consumers
on older message-format versions. Unset defers to the Kafka server
default.

### spec.config.messageFormatVersion

`string`

(Optional) Message format version the broker appends with. Consumers on
older versions may not understand newer formats; once raised it should
not be lowered. Values are the exact Kafka version tokens DigitalOcean
accepts (case-sensitive).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["0.8.0","0.8.1","0.8.2","0.9.0","0.10.0","0.10.0-IV0","0.10.0-IV1","0.10.1","0.10.1-IV0","0.10.1-IV1","0.10.1-IV2","0.10.2","0.10.2-IV0","0.11.0","0.11.0-IV0","0.11.0-IV1","0.11.0-IV2","1.0","1.0-IV0","1.1","1.1-IV0","2.0","2.0-IV0","2.0-IV1","2.1","2.1-IV0","2.1-IV1","2.1-IV2","2.2","2.2-IV0","2.2-IV1","2.3","2.3-IV0","2.3-IV1","2.4","2.4-IV0","2.4-IV1","2.5","2.5-IV0","2.6","2.6-IV0","2.7","2.7-IV0","2.7-IV1","2.7-IV2","2.8","2.8-IV0","2.8-IV1","3.0","3.0-IV0","3.0-IV1","3.1","3.1-IV0","3.2","3.2-IV0","3.3","3.3-IV0","3.3-IV1","3.3-IV2","3.3-IV3","3.4","3.4-IV0","3.5","3.5-IV0","3.5-IV1","3.5-IV2","3.6","3.6-IV0","3.6-IV1","3.6-IV2"]}}

### spec.config.messageTimestampDifferenceMaxMs

`int64` · optional (explicit presence)

(Optional) Maximum allowed difference (ms) between the broker's clock
and a message's timestamp; messages outside it are rejected when
message_timestamp_type is create_time.

### spec.config.messageTimestampType

`string`

(Optional) Which timestamp is stored on messages: the producer's
create_time or the broker's log_append_time.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["create_time","log_append_time"]}}

### spec.config.minCleanableDirtyRatio

`double` · optional (explicit presence)

(Optional) Minimum ratio of dirty (uncompacted) log to total log before
the log compactor considers the partition, 0.0 to 1.0.

- rule: {"double":{"lte":1,"gte":0}}

### spec.config.minCompactionLagMs

`uint64` · optional (explicit presence)

(Optional) Minimum time (ms) a message remains uncompacted.

### spec.config.minInsyncReplicas

`int32` · optional (explicit presence)

(Optional) Minimum number of in-sync replicas that must acknowledge a
write when the producer uses acks=all. This is the config block's only
leaf the provider defaults on the WRITE side (to 1): whenever this
config message is present and the leaf is unset, 1 is sent, even if
the server was tuned differently out-of-band. The value IS read back
from the server afterwards (measured 2026-09-17: a topic created with
2 read 2 back on every plan), so a set value never drifts.

- rule: {"int32":{"gte":1}}

### spec.config.preallocate

`bool` · optional (explicit presence)

(Optional) Preallocate a file on disk when creating a new log segment.
Unset defers to the Kafka server default.

### spec.config.retentionBytes

`int64` · optional (explicit presence)

(Optional) Maximum size (bytes) a partition retains before old segments
are discarded; -1 means no size limit.

### spec.config.retentionMs

`int64` · optional (explicit presence)

(Optional) Maximum time (ms) a log is retained before old segments are
discarded; -1 means no time limit.

### spec.config.segmentBytes

`uint64` · optional (explicit presence)

(Optional) Segment file size (bytes); retention and compaction always
act on whole segments.

### spec.config.segmentIndexBytes

`uint64` · optional (explicit presence)

(Optional) Size (bytes) of the index mapping offsets to file positions,
preallocated per segment.

### spec.config.segmentJitterMs

`uint64` · optional (explicit presence)

(Optional) Maximum random jitter (ms) subtracted from segment_ms to
avoid thundering-herd segment rolling.

### spec.config.segmentMs

`uint64` · optional (explicit presence)

(Optional) Maximum time (ms) before the log rolls to a new segment even
if the segment is not full.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanDatabaseKafkaTopic, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.cluster_id` | `string` | UUID of the Kafka database cluster the topic lives in. |
| `status.outputs.topic_name` | `string` | Name of the Kafka topic (its API identity within the cluster). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.cluster` | DigitalOceanDatabaseCluster | `status.outputs.cluster_id` |

## See Also

- [Overview](../README.md)
