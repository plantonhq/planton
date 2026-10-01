# StripeShippingRate — OpenTofu Module

OpenTofu module that declares one Stripe shipping rate. Stripe kinds run on OpenTofu only; there is no Pulumi module.

## What It Creates

- `stripe_shipping_rate` -- the rate, with `type` set to `fixed_amount`. The name, amount, currency, delivery estimate and tax code replace it; tax behavior, other currencies, `active` and `metadata` update in place. Destroy deactivates it.

## Prerequisites

- [OpenTofu](https://opentofu.org/) (the provider is pinned exactly at `stripe/stripe` `0.3.0` from OpenTofu's registry).
- A Stripe key in `STRIPE_API_KEY`: a restricted key with "Shipping Rates" write.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Cloud resource metadata (`name`, `org`, `env`, ...) |
| `spec` | `display_name` and `fixed_amount` (required); `delivery_estimate`, `tax_behavior`, `tax_code`, `active`, `metadata` |

## Outputs

| Name | Description |
|---|---|
| `id` | The shipping rate's Stripe id (`shr_...`) |
| `active` | `false` once deactivated |

## Recovering a Shipping Rate That Cannot Be Read

The provider does not treat a missing shipping rate as gone, so the next plan fails reading it. Remove it from state (`tofu state rm stripe_shipping_rate.this`) and apply again.
