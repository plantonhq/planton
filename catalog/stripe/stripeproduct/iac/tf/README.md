# StripeProduct — OpenTofu Module

OpenTofu module that declares one Stripe product and the entitlement features it grants. Stripe kinds run on OpenTofu only; there is no Pulumi module.

## What It Creates

- `stripe_product` -- the product. `type` is create-only, so changing it replaces the product. Destroy archives it. `default_price_data` is never set.
- `stripe_product_feature` -- one link per granted feature, keyed by the feature id. A link cannot be updated; removing a feature deletes its link.

## Prerequisites

- [OpenTofu](https://opentofu.org/) (the provider is pinned exactly at `stripe/stripe` `0.3.0` from OpenTofu's registry).
- A Stripe key in `STRIPE_API_KEY`: a restricted key with "Products" write, and "Entitlements" write when features are granted.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Cloud resource metadata (`name`, `org`, `env`, ...) |
| `spec` | `name` (required); `description`, `active`, `type` (replaces), `images`, `marketing_features`, `package_dimensions`, `shippable`, `statement_descriptor`, `tax_code`, `unit_label`, `url`, `metadata`, `features` (resolved feature ids) |

## Outputs

| Name | Description |
|---|---|
| `id` | The product's Stripe id (`prod_...`) |
| `active` | `false` once archived |
| `default_price` | The default price, when one is set |
| `product_feature_ids` | Each granted feature's id mapped to its link's id |

## Recovering a Product That Cannot Be Read

The provider does not treat a missing product as gone, so the next plan fails reading it. Remove it from state (`tofu state rm stripe_product.this`, and each `stripe_product_feature.this[...]`) and apply again.
