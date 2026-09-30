# Seat Management

This preset declares the capability to invite and remove teammates, the line most team plans draw between a solo tier and a paid one. The metadata records which part of your application the feature unlocks, so anyone reading the Dashboard knows what it gates.

## When to Use

- Team plans where inviting colleagues is what a customer pays for
- As a pattern for any feature whose gate lives on one route

## Key Configuration Choices

- **Lookup key** (`lookupKey`) -- `seat-management`, checked by the team settings page
- **Metadata** (`metadata.gated_route`) -- a note for your team; Stripe does not read it
- **Destroy archives** -- Stripe keeps the feature, archived

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.metadata.gated_route` | The route the feature unlocks | Your application's routing |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-api-access** -- a capability for API access
