# US Sales Tax

This preset is New York City's combined sales tax, added on top of the amount.

## When to Use

- A state or city sales tax on invoices you tax yourself
- One declared rate per jurisdiction you collect in

## Key Configuration Choices

- **Rate** (`percentage: 8.875`) -- can never change; a new rate creates a new tax rate
- **Where** (`country`, `state`, `jurisdiction`) -- the state is the ISO subdivision code without the country prefix
- **Type** (`taxType: sales_tax`) -- updates in place

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.percentage`, `spec.state`, `spec.jurisdiction` | Your jurisdiction's rate | Your tax adviser |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-eu-vat-exclusive** -- German VAT
