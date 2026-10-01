# Amazon EventBridge

This preset streams billing events into Amazon EventBridge, where rules route them to queues, functions and data pipelines without a webhook route to run. Stripe creates a partner event source in your AWS account; the destination stays pending until an event bus in that account associates with it. `status.outputs.aws_event_source_name` is the name the bus needs.

## When to Use

- Event-driven architectures already built on EventBridge
- Fanning Stripe events out to several AWS consumers

## Key Configuration Choices

- **Where** (`amazonEventbridge.awsAccountId`, `awsRegion`) -- changing either replaces the destination and its event source
- **Association** -- create an AwsEventBridgeBus with `eventSourceName` set to `status.outputs.aws_event_source_name`
- **Import replaces it** -- Stripe's API does not return the region, so an imported EventBridge destination is replaced on its first apply
- **Destroy deletes** -- the destination and its event source

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.amazonEventbridge.awsAccountId` | The AWS account that receives events | AWS console -> account menu |
| `spec.amazonEventbridge.awsRegion` | The region of the event bus | Your AWS architecture |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-thin-events-webhook** -- thin events to a webhook route
