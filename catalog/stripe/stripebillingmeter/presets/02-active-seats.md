# Active Seats

This preset bills the last seat count your application reported in each period: send an `active_seats` event with a `seats` value whenever a customer's seat count changes.

## When to Use

- Per-seat billing where the count changes during the month
- Any usage billed at its latest value rather than its total

## Key Configuration Choices

- **Last value** (`formula: last`) -- the period's last reported value is billed
- **Value key** (`valueSettings.eventPayloadKey: seats`) -- where Stripe reads the count in each event
- **Customer** -- left unset, so Stripe reads `stripe_customer_id` from each event

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.eventName` | The event your application sends | Your usage-reporting code |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-api-requests-with-alert** -- a summed meter with a usage alert
