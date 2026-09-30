# Graduated Usage

This preset bills API calls each month in graduated tiers: the first ten thousand are free, the next million cost a tenth of a cent each, and everything above costs half that. Usage is reported to a StripeBillingMeter, named by reference, and the price bills whatever the meter counted in the period.

## When to Use

- Usage-based products where volume customers earn a lower rate
- A metered add-on beside a flat subscription price

## Key Configuration Choices

- **Tiers** (`billingScheme: tiered`, `tiersMode: graduated`, `tiers`) -- each unit is billed at the tier it falls in; the last tier is `inf`
- **Sub-cent rates** (`unitAmountDecimal`) -- up to 12 decimal places of a cent
- **Metering** (`recurring.usageType: metered`, `meter`) -- a metered price names its meter; a replaced meter replaces the price with it
- **Replacement** -- tiers can never change on a price; a new tier table creates a new price and archives the old one

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.product.valueFrom.name` | The name of your StripeProduct | Its manifest's `metadata.name` |
| `spec.recurring.meter.valueFrom.name` | The name of your StripeBillingMeter | Its manifest's `metadata.name` |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-monthly-subscription** -- a flat monthly price in three currencies
