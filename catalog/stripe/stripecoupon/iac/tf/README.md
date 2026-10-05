# StripeCoupon — OpenTofu Module

OpenTofu module that declares one Stripe coupon. Stripe kinds run on OpenTofu only; there is no Pulumi module.

## What It Creates

- `stripe_coupon` -- the coupon. Only `name` and `metadata` update in place; any other change replaces it (the old coupon is deleted, the new one created). Destroy deletes it; customers who already applied it keep their discount.
- `terraform_data.replace_triggers` -- holds the products the coupon applies to and its amounts in other currencies, which Stripe never returns to the provider's read. The coupon ignores both after create, and a change to either replaces it through this tracker, so an imported coupon is adopted untouched.

## Prerequisites

- [OpenTofu](https://opentofu.org/) (the provider is pinned exactly at `stripe/stripe` `0.3.0` from OpenTofu's registry).
- A Stripe key in `STRIPE_API_KEY`: a restricted key with "Coupons" write.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Catalog object metadata (`name`, `org`, `env`, ...) |
| `spec` | `percent_off`, or `amount_off` with `currency` (exactly one); `name`, `duration`, `duration_in_months`, `max_redemptions`, `redeem_by`, `applies_to_products` (resolved product ids), `currency_options`, `metadata` |

## Outputs

| Name | Description |
|---|---|
| `id` | The coupon's Stripe id |
| `valid` | Whether new customers can still redeem it |

## Recovering a Coupon That Cannot Be Read

The provider does not treat a missing coupon as gone, so the next plan fails reading it. Remove it from state (`tofu state rm stripe_coupon.this`) and apply again.
