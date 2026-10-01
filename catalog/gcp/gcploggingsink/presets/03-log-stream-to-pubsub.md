# Log Stream to Pub/Sub

The front door to third-party log pipelines: entries stream to a
Pub/Sub topic in near real time, where Datadog/Splunk-class collectors
(or your own subscribers) consume them.

## What it configures

- A `pubsubTopic` destination — the module renders the
  `pubsub.googleapis.com/...` URI from the full topic path.
- An EMPTY main filter (everything streams) with a sampling exclusion
  that drops 90% of DEBUG noise — the volume/completeness trade made
  explicit.

## The deploy's second half

Grant the sink's `writer_identity` output `roles/pubsub.publisher` on
the topic — through a `GcpPubSubTopicIamMember` in the same chart whose
`member` references the sink's `status.outputs.writer_identity`.

## Adjust before deploying

- **pubsubTopic** — reference a GcpPubSubTopic resource via valueFrom
  (its `topic_id` output is the full path this field wants).
- Tune the exclusion sampling to your pipeline's ingest pricing; add
  more carve-outs per noisy service.

## When to choose something else

If the consumer is a human with SQL, the **Audit Logs to BigQuery**
preset removes the pipeline entirely; if it is a compliance archive,
GCS is cheaper than any stream.
