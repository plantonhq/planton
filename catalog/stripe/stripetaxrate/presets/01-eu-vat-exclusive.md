# EU VAT, Exclusive

This preset is German VAT at the standard 19%, added on top of the amount, for accounts that apply tax rates themselves instead of using Stripe Tax.

## When to Use

- Invoices and subscriptions that list VAT as its own line
- One declared rate per country you sell in

## Key Configuration Choices

- **Rate** (`percentage: 19`) -- can never change; a new rate creates a new tax rate, and subscriptions keep the old one until you move them
- **Exclusive** (`inclusive: false`) -- tax is added on top; can never change
- **Where** (`country`, `jurisdiction`) -- update in place
- **Destroy deactivates** -- it still applies to subscriptions and invoices that already use it

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.percentage`, `spec.country` | Your country's rate | Your tax adviser |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-us-sales-tax** -- a state sales tax
