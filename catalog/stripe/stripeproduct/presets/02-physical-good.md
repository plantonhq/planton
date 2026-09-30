# Physical Good

This preset declares a shippable product with its package dimensions, so shipping rates and labels can be computed from the product itself. A good carries no statement descriptor or unit label -- Stripe allows those only on a service.

## When to Use

- Hardware, merchandise, or anything that leaves a warehouse
- One-time purchases through Checkout with shipping address collection

## Key Configuration Choices

- **Type** (`type: good`) -- sent only when the product is created; changing it replaces the product
- **Shipping** (`shippable`, `packageDimensions`) -- inches and ounces, as Stripe records them
- **Images** (`images`) -- up to eight public URLs shown to customers
- **Destroy archives** -- the product can no longer be bought; Stripe keeps it

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.images` | Public URLs of product photos | Your CDN or storefront |
| `spec.packageDimensions` | The boxed product's size and weight | Your fulfillment records |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-saas-plan** -- a subscription plan that grants entitlement features
