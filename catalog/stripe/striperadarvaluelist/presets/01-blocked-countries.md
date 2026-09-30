# Blocked Countries

This preset declares a list of countries a Radar rule blocks, such as "Block if :ip_country: in @blocked_countries". The list holds the values; the rule, written in the Dashboard, holds the decision. Items are checked as two-letter country codes before Stripe sees them.

## When to Use

- Compliance or fraud policies that refuse payments from certain countries
- Any Radar rule whose values change more often than its logic

## Key Configuration Choices

- **Alias** (`alias`) -- the name rules use after `@`; renaming it means changing every rule that names it
- **Type** (`itemType: country`) -- changing it replaces the list and every item
- **Items** -- adding one creates it, removing one deletes it
- **Cost** -- the list is free; custom rules that use it need Radar for Fraud Teams, priced per screened transaction

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.items` | The countries your policy blocks | Your compliance policy |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-trusted-customers** -- customers a rule lets through
