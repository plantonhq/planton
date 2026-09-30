# Thin Events Webhook

This preset delivers thin events -- small notifications that name what happened, which your service fetches details for -- to one webhook route. Thin events are how Stripe's v2 APIs report, and they keep deliveries small and version-independent. The signing secret is captured at creation into `status.outputs.signing_secret`, for your service to read by reference.

## When to Use

- A service that handles v2 events, such as billing meter errors
- Any receiver that prefers fetching current state over parsing a snapshot

## Key Configuration Choices

- **Payload** (`eventPayload: thin`) -- changing it replaces the destination and rotates its secret
- **Events** (`enabledEvents`) -- only what the route handles
- **Signing secret** (`status.outputs.signing_secret`) -- exists only at creation; an imported destination has none
- **Destroy deletes** -- the destination and its secret are gone

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.webhookEndpoint.url` | Your service's thin-event route | Your application's routing |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-amazon-eventbridge** -- snapshot events into an AWS event bus
