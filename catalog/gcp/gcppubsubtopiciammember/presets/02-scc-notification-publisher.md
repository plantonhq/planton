# Security Command Center Publisher

This preset grants `roles/pubsub.publisher` on one topic to the service account a Security Command Center notification config publishes as. Without it, findings matching the config never reach the topic.

## When to Use

- A `GcpSccNotificationConfig` streams findings to a `GcpPubSubTopic`
- Routing security findings into a ticketing or chat pipeline that subscribes to the topic
- Keeping the findings publisher scoped to exactly its own topic

## Key Configuration Choices

- **`valueFrom` member** — the notification config's `service_account_member` output is the publishing identity in member format; no email assembly needed
- **Topic-scoped** — the publisher reaches only the findings topic, never the project's other streams
- **Same chart as the notification config** — the grant lands in the same deploy, so the first finding is delivered

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<topic-resource-name>` | The Planton resource name of the findings GcpPubSubTopic | Your GcpPubSubTopic manifest's `metadata.name` |
| `<scc-notification-config-resource-name>` | The Planton resource name of the GcpSccNotificationConfig | Your GcpSccNotificationConfig manifest's `metadata.name` |

## Related Presets

- **01-logging-sink-publisher** — The same grant for a logging sink's writer identity
- **03-workload-subscriber** — Let a workload attach subscriptions to the topic
