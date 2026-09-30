# StripePromotionCode — OpenTofu Module

OpenTofu module that declares one Stripe promotion code on a coupon. Stripe kinds run on OpenTofu only; there is no Pulumi module.

## What It Creates

- `stripe_promotion_code` -- the code, with `promotion.type` set to `coupon`. Only `active` and `metadata` update in place; any other change replaces it (the old code is deactivated first, so the same code can be reused). Destroy deactivates it.

## Prerequisites

- [OpenTofu](https://opentofu.org/) (the provider is pinned exactly at `stripe/stripe` `0.3.0` from OpenTofu's registry).
- A Stripe key in `STRIPE_API_KEY`: a restricted key with "Promotion Codes" write.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Cloud resource metadata (`name`, `org`, `env`, ...) |
| `spec` | `coupon` (resolved coupon id, required); `code`, `customer`, `customer_account`, `expires_at`, `max_redemptions`, `restrictions`, `active`, `metadata` |

## Outputs

| Name | Description |
|---|---|
| `id` | The promotion code's Stripe id (`promo_...`) |
| `code` | What customers type |
| `active` | `false` once deactivated |

## Recovering a Promotion Code That Cannot Be Read

The provider does not treat a missing promotion code as gone, so the next plan fails reading it. Remove it from state (`tofu state rm stripe_promotion_code.this`) and apply again.
