# 25% Off for Three Months

This preset takes 25% off a subscription's first three months, on one product only, for at most 500 customers. Customers redeem it with a promotion code (declare a StripePromotionCode that names this coupon), or your application applies it to a subscription or Checkout session.

## When to Use

- A launch or seasonal discount on a plan
- A win-back offer for a single product

## Key Configuration Choices

- **Discount** (`percentOff: 25`) -- can never change; a new percentage creates a new coupon and replaces its promotion codes
- **Duration** (`duration: repeating`, `durationInMonths: 3`) -- three monthly invoices, then full price
- **Products** (`appliesToProducts`) -- by reference, so the coupon follows the product declared beside it
- **Destroy deletes** -- customers who already applied it keep their discount; nobody new can redeem it

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.appliesToProducts[0].valueFrom.name` | The name of your StripeProduct | Its manifest's `metadata.name` |
| `spec.name` | What customers see on invoices | Your campaign |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-ten-off-once** -- a fixed amount off the first invoice, in two currencies
