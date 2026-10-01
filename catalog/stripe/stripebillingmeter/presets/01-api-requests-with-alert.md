# API Requests With an Alert

This preset counts API requests per customer: your application sends an `api_requests` event with the customer's id and a count, the meter sums them each billing period, and Stripe sends a `billing.alert.triggered` event when a customer reaches 10,000. A metered StripePrice names this meter to bill what it counts.

## When to Use

- Usage-based pricing on requests, messages, or any countable unit
- An early warning when a customer's usage crosses a threshold

## Key Configuration Choices

- **Event name** (`eventName: api_requests`) -- your application reads it from `status.outputs.event_name` by reference, so the name it sends and the name the meter counts never drift; changing it creates a new meter
- **Sum** (`formula: sum`) -- adds up each event's `value`
- **Alert** (`alerts`) -- can't be changed; a change creates a new alert and leaves the old one active, and removing it leaves it active too
- **Destroy deactivates** -- the meter stops counting; its alerts stay active until archived

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.eventName` | The event your application sends | Your usage-reporting code |
| `spec.alerts` | The thresholds you want to hear about | Your pricing tiers |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-active-seats** -- the latest seat count, for per-seat usage billing
