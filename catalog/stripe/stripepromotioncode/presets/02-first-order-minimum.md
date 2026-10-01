# First Order Minimum

This preset is a welcome code that applies only to a customer's first order of at least 50 dollars (45 euros for customers paying in euros).

## When to Use

- A new-customer discount that should not apply to small orders
- A code shared with a partner audience

## Key Configuration Choices

- **Minimum** (`minimumAmount`, `minimumAmountCurrency`) -- the order total the code needs, in cents
- **One currency** -- the minimum is set in the main currency only. Minimums in other currencies (`currencyOptions`) can't be declared yet: the pinned Stripe provider can't hold them, so validation refuses them
- **First-time customers** (`firstTimeTransaction`) -- customers who have never paid before
- **Changing a limit replaces the code** -- the old code is deactivated first, so the same code keeps working

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.coupon.valueFrom.name` | The name of your StripeCoupon | Its manifest's `metadata.name` |
| `spec.code` | What customers type | Your campaign |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-launch-code** -- a capped, expiring launch code
