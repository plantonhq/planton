# SaaS Plan

This preset declares a subscription plan the way customers meet it -- a name, a description, three lines in your pricing table -- and the capabilities buying it grants, by reference to the entitlement features declared beside it. Its prices are separate StripePrice resources that name this product.

## When to Use

- A SaaS plan sold through Checkout, Payment Links or a pricing table
- Plans whose capabilities your application checks as entitlements

## Key Configuration Choices

- **What customers read** (`name`, `description`, `marketingFeatures`) -- shown on Stripe-hosted pages and the pricing table
- **What a purchase grants** (`features`) -- entitlement features by reference; adding or removing one changes every subscriber's entitlements
- **Tax category** (`taxCode`) -- `txcd_10103001` is software as a service; Stripe Tax uses it when Tax is on
- **Seats** (`unitLabel`) -- receipts and invoices say "seat"
- **Destroy archives** -- the product can no longer be bought; existing subscriptions keep billing

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.name`, `spec.description` | What customers see | Your pricing page |
| `spec.features[].valueFrom.name` | The names of your StripeEntitlementFeature resources | Their manifests' `metadata.name` |
| `spec.statementDescriptor` | The text on card statements, up to 22 characters | Your brand |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-physical-good** -- a shippable product with package dimensions
