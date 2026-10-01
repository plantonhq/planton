# Project High Findings

## Use Case

Page the team that owns a project when Security Command Center finds an active high or critical problem in it.

## When to Use

- A project team that owns its own security response
- Feeding a project's findings into an existing alerting pipeline

## What This Creates

- The Security Command Center API on the project
- A notification config streaming active, unmuted high and critical findings to the topic

Grant the config's `service_account` output `roles/pubsub.publisher` on the topic.

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `pubsubTopic` | `scc-findings` | Reference your `GcpPubSubTopic`. |
| `filter` | active, unmuted HIGH/CRITICAL | Add categories, or drop `NOT mute = "MUTED"` to stream muted findings too. |
