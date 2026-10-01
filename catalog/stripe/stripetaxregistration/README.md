# StripeTaxRegistration

Declares one place your account is registered to collect tax with Stripe Tax -- VAT in Germany under the EU's One-Stop Shop, sales tax in Texas -- so your registrations live in files, the same in every environment. See Stripe's [tax registrations](https://docs.stripe.com/tax/registering).

## When to Use

- **Your registrations, on file**: one resource per place you collect tax, reviewed like code.
- **Scheduled changes**: a registration that starts on a future date, or one that expires when you deregister.

Stripe Tax itself (turning it on, your origin address, your default tax code) is set in the Stripe Dashboard, not here.

## Quick Start

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeTaxRegistration
metadata:
  name: de-oss
  org: acme-corp
  env: production
spec:
  country: DE
  type: oss_union
  activeFrom: 1830297600  # your start: now or later, at most five years ahead, in Unix seconds
```

## Fields

| Field | Description |
|---|---|
| `country` | Two-letter country code in capitals (required). **Replaces** |
| `type` | The kind of registration; the country decides which types exist (required). **Replaces** |
| `activeFrom` | When it starts, in Unix seconds (required); now or later when it is created. Changes in place |
| `expiresAt` | When it stops, in Unix seconds. Changes in place, and removing it does not clear it |
| `placeOfSupplyScheme` | `standard`, `inbound_goods`, or `small_seller` (EU only), for a standard registration. **Replaces** |
| `province` | A Canadian province, for `province_standard`. **Replaces** |
| `state` | A US state, required for US registrations. **Replaces** |
| `jurisdiction` | A FIPS code, for `local_amusement_tax` and `local_lease_tax`. **Replaces** |
| `stateSalesTaxElections` | Local-tax elections of a `state_sales_tax` registration. **Replaces** |

## Key Behaviors

- **Applying starts collection.** From `activeFrom`, Stripe Tax collects tax in that place on every payment that uses automatic tax.
- **Destroy only forgets.** Stripe never deletes a registration and keeps collecting. Set `expiresAt` and apply to stop.
- **Only the dates change in place.** Any other change creates a new registration, and the old one keeps collecting until it expires.
- **Runs on OpenTofu only.** Planton refuses any other engine for this kind before anything runs.

## Outputs

| Output | Description |
|---|---|
| `id` | The registration's Stripe id (`taxreg_...`); it changes when the registration is replaced |
| `status` | `scheduled`, `active` or `expired` |

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
