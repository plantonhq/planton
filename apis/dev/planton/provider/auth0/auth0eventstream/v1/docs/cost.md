# Auth0 Event Stream - Cost

## Pricing Model

Auth0 pricing is based on Monthly Active Users (MAUs), not on the number of resources created. Event streams (log streams) are free API objects with no per-resource cost.

## Free Tier

The Auth0 Free plan includes:

- 25,000 MAUs
- 1 tenant
- Log streams available on all plans

## Cost Impact

Creating, updating, or deleting Auth0 event stream resources has no direct billing impact from Auth0. There is no charge per log stream definition.

The only Auth0 cost driver is the number of monthly active users authenticating through your tenant.

## Downstream Cost Considerations

While Auth0 does not charge for event streams, the destination services may incur costs:

| Destination Type | Cost Consideration |
|-----------------|-------------------|
| Webhook (HTTP) | Cost depends on your endpoint's hosting |
| Amazon EventBridge | EventBridge charges per event published |
| Datadog | Log ingestion charges apply |
| Splunk | Log volume-based pricing applies |
| Sumo Logic | Ingestion-based pricing applies |
| Azure Event Hubs | Throughput unit pricing applies |

## Event Volume

Event volume scales with authentication activity, not with the number of log streams. Creating multiple streams to the same destination type multiplies the downstream ingestion cost but not the Auth0 cost.

## Log Retention

Auth0's built-in log retention is limited by plan tier (2 days free, 30 days enterprise). Event streams provide a mechanism to export logs to external systems for longer retention without affecting Auth0 costs.
