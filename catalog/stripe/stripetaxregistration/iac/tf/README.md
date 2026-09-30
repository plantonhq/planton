# StripeTaxRegistration — OpenTofu Module

OpenTofu module that declares one Stripe Tax registration. Stripe kinds run on OpenTofu only; there is no Pulumi module.

## What It Creates

- `stripe_tax_registration` -- the registration. The module writes the one `country_options` block the declared country takes (`country_options.de`, `country_options.us`, ...), with only the parts the declared type uses. `active_from` and `expires_at` update in place; every other change replaces the registration and forgets the old one while it is still active. Destroy only removes it from state; Stripe keeps collecting until `expires_at`.

## Prerequisites

- [OpenTofu](https://opentofu.org/) (the provider is pinned exactly at `stripe/stripe` `0.3.0` from OpenTofu's registry).
- A Stripe key in `STRIPE_API_KEY`: a restricted key with "Tax Registrations" write.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Cloud resource metadata (`name`, `org`, `env`, ...) |
| `spec` | `country`, `type`, `active_from` (required); `expires_at`, `place_of_supply_scheme`, `province`, `state`, `jurisdiction`, `state_sales_tax_elections` |

## Outputs

| Name | Description |
|---|---|
| `id` | The registration's Stripe id (`taxreg_...`) |
| `status` | `scheduled`, `active` or `expired` |

## Where the Country Rules Come From

The spec's validation lists which countries take which types, and where a place-of-supply scheme, a province, a state or a jurisdiction belongs. Those sets are the pinned provider's own `country_options` schema (101 countries in five shapes). A provider version move re-reads them.

## Recovering a Registration That Cannot Be Read

The provider does not treat a missing registration as gone, so the next plan fails reading it. Remove it from state (`tofu state rm stripe_tax_registration.this`) and apply again. A new registration needs a start date that is now or later.
