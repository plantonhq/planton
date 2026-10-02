# Logging Sink Publisher

This preset grants `roles/pubsub.publisher` on one topic to a logging sink's writer identity — the grant every Pub/Sub log export needs before Cloud Logging can deliver a single entry. Both sides arrive by reference, so the grant deploys after the topic and the sink without a dependency cycle.

## When to Use

- A `GcpLoggingSink` exports to a `GcpPubSubTopic` (streaming logs to a SIEM or a third-party pipeline)
- Any Pub/Sub log export deployed in one chart, where the export must work from its first entry
- Replacing a project-wide publisher grant with one scoped to the sink's own destination

## Key Configuration Choices

- **`valueFrom` member** — the sink's `writer_identity` output is minted by Google; the reference keeps the grant on the current identity even if the sink is recreated
- **`valueFrom` topic** — the topic's `topic_id` output carries the project, so the grant needs no project field
- **Publisher only** — the writer can publish to this one topic and read nothing

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<topic-resource-name>` | The Planton resource name of the destination GcpPubSubTopic | Your GcpPubSubTopic manifest's `metadata.name` |
| `<logging-sink-resource-name>` | The Planton resource name of the GcpLoggingSink exporting to the topic | Your GcpLoggingSink manifest's `metadata.name` |

## Related Presets

- **02-scc-notification-publisher** — The same grant for Security Command Center notifications
- **03-workload-subscriber** — Let a workload attach subscriptions to the topic
