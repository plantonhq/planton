# EU One-Stop Shop

This preset is a registration for the EU's One-Stop Shop (union scheme), declared in Germany, the country the account registered it in. It lets Stripe Tax collect VAT on sales to consumers across the EU.

## When to Use

- A business selling digital services or goods to EU consumers, registered for OSS in its home member state
- One registration covers every EU country; add a `standard` registration for each country where you are also registered locally

## Key Configuration Choices

- **Country** (`country: DE`) -- the member state where you registered for OSS; changing it creates a new registration
- **Type** (`type: oss_union`) -- the union scheme; `oss_non_union` for businesses established outside the EU, `ioss` for imported goods
- **Start** (`activeFrom`) -- the preset's date is 1 January 2028, a placeholder: applying it unchanged schedules a registration that starts collecting tax then. Set your real start: now or later, at most five years ahead, in Unix seconds
- **Destroy only forgets** -- Stripe keeps collecting; set `expiresAt` and apply to stop

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.country` | The member state you registered OSS in | Your OSS registration |
| `spec.activeFrom` | When collection starts, now or later | `date -u -d <date> +%s` on Linux |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-us-state-sales-tax** -- a US state sales tax registration
