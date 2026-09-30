# API Access

This preset declares one capability -- access to your API -- that the plans you sell can grant. Your application checks a customer's active entitlements for the lookup key `api-access` instead of hard-coding which prices include the API, so moving the capability between plans never touches code.

## When to Use

- A SaaS whose plans differ by what they unlock
- Any capability your code gates, declared once and granted by as many products as need it

## Key Configuration Choices

- **Lookup key** (`lookupKey`) -- the stable name your code checks; changing it replaces the feature, so choose it once
- **Name** (`name`) -- your team's label in the Dashboard, never shown to customers
- **Destroy archives** -- Stripe keeps the feature, archived; it can no longer be attached to new products

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.lookupKey` | The name your application checks | Your application's entitlement checks |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-seat-management** -- a capability for managing team seats
