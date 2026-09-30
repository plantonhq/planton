# Trusted Customers

This preset declares customers a Radar rule lets through without review, such as "Allow if :customer: in @trusted_customers" -- long-standing accounts whose large orders would otherwise be flagged. Items are checked as Stripe customer ids before Stripe sees them.

## When to Use

- Enterprise customers whose order sizes look unusual to fraud rules
- Any allow list keyed by customer

## Key Configuration Choices

- **Type** (`itemType: customer_id`) -- each item must be a `cus_...` id
- **Items** -- adding one creates it, removing one deletes it
- **Destroy deletes** -- Stripe refuses while a rule still uses the list; remove the rule first

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.items` | Your trusted customers' ids | Stripe Dashboard -> Customers |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-blocked-countries** -- countries a rule blocks
