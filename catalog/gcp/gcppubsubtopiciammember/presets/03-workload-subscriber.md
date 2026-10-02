# Workload Subscriber

This preset grants `roles/pubsub.subscriber` on one topic to a workload's service account, so the workload can attach its own subscriptions to the topic — including from another project.

## When to Use

- A consumer team creates subscriptions on a topic another team owns
- A cross-project pipeline reads a stream published in a central project
- Both the topic and the consumer identity are Planton-managed and the access edge should be visible in the graph

## Key Configuration Choices

- **Subscriber, not publisher** — the workload can attach and consume; it cannot publish into the topic
- **Both sides referenced** — the topic's `topic_id` and the account's `member` keep the whole relationship in the resource graph
- **Topic-scoped** — a project-level subscriber grant would expose every topic in the project

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<topic-resource-name>` | The Planton resource name of the GcpPubSubTopic | Your GcpPubSubTopic manifest's `metadata.name` |
| `<service-account-resource-name>` | The consuming workload's GcpServiceAccount | Your GcpServiceAccount manifest's `metadata.name` |

## Related Presets

- **01-logging-sink-publisher** — Let a logging sink publish into the topic
- **02-scc-notification-publisher** — Let Security Command Center publish findings into the topic
