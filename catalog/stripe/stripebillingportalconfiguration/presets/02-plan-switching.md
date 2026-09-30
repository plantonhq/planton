# Plan Switching Portal

This preset lets customers move between the plans you sell -- monthly and yearly, more or fewer seats -- on their own. An upgrade applies at once and credits the unused time to the next invoice; a change that lowers what they pay waits until the period ends, so nobody is refunded mid-period.

## When to Use

- A product sold in tiers or seats where customers should self-serve upgrades and downgrades
- Alongside cancellation at period end, for a portal that handles every plan change

## Key Configuration Choices

- **What may change** (`subscriptionUpdate.defaultAllowedUpdates`) -- the price and the quantity; promotion codes stay with your checkout
- **Plans offered** (`subscriptionUpdate.products`) -- each product with the prices a customer may choose, and a seat range; Stripe allows up to ten products. They reference the StripeProduct and StripePrice resources declared beside the portal, so a price replaced after an amount change reaches the portal on its next apply
- **Proration** (`prorationBehavior: create_prorations`) -- an upgrade's difference lands on the next invoice
- **Downgrades wait** (`scheduleAtPeriodEnd` with `decreasing_item_amount`) -- a cheaper plan starts at period end
- **Destroy deactivates** -- Stripe keeps the configuration, inactive, forever

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.features.subscriptionUpdate.products[].product` | The name of your StripeProduct (`team-plan`), or a literal `value: prod_...` for a product declared elsewhere | Your StripeProduct manifest's `metadata.name`, or Stripe Dashboard -> Product catalog |
| `spec.features.subscriptionUpdate.products[].prices` | The names of its StripePrice resources (`team-monthly`, `team-yearly`), or literal `value: price_...` ids | Your StripePrice manifests, or the product's page, Pricing section |
| `spec.defaultReturnUrl` | Your application's billing page | Your application's routing |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-self-serve-at-period-end** -- details, payment methods, invoices and cancellation, without plan changes
