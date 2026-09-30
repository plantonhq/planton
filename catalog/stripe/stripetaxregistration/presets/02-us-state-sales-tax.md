# US State Sales Tax

This preset is a state sales tax registration in Texas. It lets Stripe Tax collect Texas sales tax on payments that use automatic tax.

## When to Use

- A business registered to collect sales tax in a US state
- One resource per state you collect in

## Key Configuration Choices

- **State** (`state: TX`) -- the state you are registered in; changing it creates a new registration
- **Type** (`type: state_sales_tax`) -- the state's sales tax; `state_communications_tax` and `state_retail_delivery_fee` are the other state-level types, and `local_amusement_tax` and `local_lease_tax` take a `jurisdiction`
- **Start** (`activeFrom`) -- the preset's date is 1 January 2100, so applying it unchanged schedules a registration that never starts in practice. Set your real start: now or later, in Unix seconds
- **Destroy only forgets** -- Stripe keeps collecting; set `expiresAt` and apply to stop

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.state` | The state you are registered in | Your state registration |
| `spec.activeFrom` | When collection starts, now or later | `date -u -d <date> +%s` on Linux |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-eu-one-stop-shop** -- an EU One-Stop Shop registration
