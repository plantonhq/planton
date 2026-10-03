# StripeTaxRate — OpenTofu Module

OpenTofu module that declares one Stripe tax rate. Stripe kinds run on OpenTofu only; there is no Pulumi module.

## What It Creates

- `stripe_tax_rate` -- the rate. `percentage` and `inclusive` replace it; everything else updates in place. Destroy deactivates it; it still applies wherever it is already used.

## Prerequisites

- [OpenTofu](https://opentofu.org/) (the provider is pinned exactly at `stripe/stripe` `0.3.0` from OpenTofu's registry).
- A Stripe key in `STRIPE_API_KEY`: a restricted key with "Tax Rates" write.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Catalog object metadata (`name`, `org`, `env`, ...) |
| `spec` | `display_name` (required), `percentage`, `inclusive`; `country`, `state`, `jurisdiction`, `description`, `tax_type`, `active`, `metadata` |

## Outputs

| Name | Description |
|---|---|
| `id` | The tax rate's Stripe id (`txr_...`) |
| `active` | `false` once deactivated |

## Recovering a Tax Rate That Cannot Be Read

The provider does not treat a missing tax rate as gone, so the next plan fails reading it. Remove it from state (`tofu state rm stripe_tax_rate.this`) and apply again.
