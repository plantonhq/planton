---
title: "Partner Event Bus"
description: "This preset creates an EventBridge custom event bus for receiving events from a SaaS partner integration. When a partner event source is configured in your AWS account, you create a bus with the same..."
type: "preset"
rank: "03"
presetSlug: "03-partner-event-bus"
componentSlug: "eventbridge-bus"
componentTitle: "EventBridge Bus"
provider: "aws"
icon: "package"
order: 3
---

# Partner Event Bus

This preset creates an EventBridge custom event bus for receiving events from a SaaS partner integration. When a partner event source is configured in your AWS account, you create a bus with the same name to start receiving partner events.

## When to Use

- Integrating with a SaaS partner that supports EventBridge (Datadog, PagerDuty, Zendesk, Auth0, etc.)
- Receiving operational events from third-party services for automated incident response
- Building event-driven workflows triggered by external SaaS events

## Key Configuration Choices

- **Event source name** (`eventSourceName`) — the full partner event source name as provided by the SaaS partner integration
- **Bus name matches event source** (`metadata.name`) — must exactly match the `eventSourceName` value; this is an AWS requirement for partner event buses

## Placeholders to Replace

| Placeholder | Description | Example |
|-------------|-------------|---------|
| `aws.partner/example.com/tenant123/orders` | Replace both `metadata.name` and `spec.eventSourceName` with the actual partner event source name from your AWS console | `aws.partner/datadog.com/12345/events` |

## Important Notes

- The `eventSourceName` is immutable — changing it forces bus replacement
- The bus name (`metadata.name`) must match `eventSourceName` exactly
- Partner event sources must be configured in your AWS account before creating the bus
- After the bus is created, you need to associate the partner event source with the bus in the AWS console or via API

## Common Additions

- Add `deadLetterConfig` to catch partner events that fail delivery to rule targets
- Add `logConfig` with level `INFO` to monitor incoming partner events
- Create EventBridge rules (AwsEventBridgeRule) to route partner events to your targets

## Related Presets

- **01-simple-custom-bus** — use for custom application events (not partner integrations)
- **02-production-encrypted-bus** — use for production workloads requiring encryption and DLQ
