# StripePrice — OpenTofu Module

OpenTofu module that declares one Stripe price on a product. Stripe kinds run on OpenTofu only; there is no Pulumi module.

## What It Creates

- `stripe_price` -- the price, with `create_before_destroy`, so a replacement (any change to the amount, currency, product, scheme, tiers or schedule) creates the new price before the old one is archived. Destroy archives it. `product_data` is never set.

## Prerequisites

- [OpenTofu](https://opentofu.org/) (the provider is pinned exactly at `stripe/stripe` `0.3.0` from OpenTofu's registry).
- A Stripe key in `STRIPE_API_KEY`: a restricted key with "Prices" write.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Cloud resource metadata (`name`, `org`, `env`, ...) |
| `spec` | `product` (resolved product id), `currency` (required); `unit_amount`, `unit_amount_decimal`, `billing_scheme`, `tiers_mode`, `tiers`, `custom_unit_amount`, `transform_quantity`, `recurring`, `currency_options`, `lookup_key`, `transfer_lookup_key`, `nickname`, `tax_behavior`, `active`, `metadata` |

## Outputs

| Name | Description |
|---|---|
| `id` | The price's Stripe id (`price_...`) |
| `type` | `one_time` or `recurring` |
| `active` | `false` once archived |
| `lookup_key` | The stable name, when set |
| `product` | The product's id |

## Recovering a Price That Cannot Be Read

The provider does not treat a missing price as gone, so the next plan fails reading it. Remove it from state (`tofu state rm stripe_price.this`) and apply again.
