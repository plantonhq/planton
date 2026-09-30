# StripeTaxRate

Declares a manual tax rate -- VAT at 19% in Germany, a state sales tax -- that invoices, subscriptions, Checkout sessions and payment links apply when you calculate tax yourself rather than with Stripe Tax. See Stripe's [tax rates](https://docs.stripe.com/billing/taxes/tax-rates).

## When to Use

- **The rates you charge, in files**: one per country or jurisdiction, the same in every environment.
- **Manual tax without Stripe Tax**: rates reviewed like code instead of typed into the Dashboard.

Declare a tax rate here only if nothing else owns it.

## Quick Start

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeTaxRate
metadata:
  name: de-vat
  org: acme-corp
  env: production
spec:
  displayName: VAT
  percentage: 19
  inclusive: false
  country: DE
  taxType: vat
```

## Fields

| Field | Description |
|---|---|
| `displayName` | The tax's name on invoices (required). Changes in place |
| `percentage` | The rate, 0 to 100. **Replaces** |
| `inclusive` | Already included in the amount (`true`) or added on top (`false`). **Replaces** |
| `country`, `state`, `jurisdiction` | Where it applies. Change in place |
| `taxType` | `vat`, `sales_tax`, `gst` and the other kinds the provider accepts. Changes in place |
| `description` | A note for your team. Changes in place |
| `active` | Whether it can be added to new invoices and subscriptions (default `true`) |
| `metadata` | Key-value pairs stored on the rate |

## Key Behaviors

- **A rate's percentage and inclusiveness never change in Stripe.** Changing either deactivates the old rate and creates a new one. Subscriptions keep the old rate until you move them.
- **Destroy deactivates the rate.** It still applies to the subscriptions and invoices that already use it.
- **Runs on OpenTofu only.** Planton refuses any other engine for this kind before anything runs.

## Outputs

| Output | Description |
|---|---|
| `id` | The tax rate's Stripe id (`txr_...`); it changes when the rate is replaced |
| `active` | `false` once deactivated |
