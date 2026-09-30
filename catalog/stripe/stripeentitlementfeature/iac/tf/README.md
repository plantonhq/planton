# StripeEntitlementFeature — OpenTofu Module

OpenTofu module that declares one Stripe Entitlements feature. Stripe kinds run on OpenTofu only; there is no Pulumi module.

## What It Creates

- `stripe_entitlements_feature` -- a feature in the account the provider's key belongs to (or, with `STRIPE_ACCOUNT`, another account it may act on). `lookup_key` is create-only, so changing it replaces the feature. Destroy archives it.

## Prerequisites

- [OpenTofu](https://opentofu.org/) (the provider is pinned exactly at `stripe/stripe` `0.3.0` from OpenTofu's registry).
- A Stripe key in `STRIPE_API_KEY`: a restricted key with "Entitlements" write.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Cloud resource metadata (`name`, `org`, `env`, ...) |
| `spec` | `lookup_key` (required, replaces), `name` (required); `metadata` (optional) |

## Outputs

| Name | Description |
|---|---|
| `id` | The feature's Stripe id (`feat_...`) |
| `lookup_key` | The stable name |
| `active` | `false` once archived |

## Recovering a Feature That Cannot Be Read

The provider does not treat a missing feature as gone, so the next plan fails reading it. Remove it from state (`tofu state rm stripe_entitlements_feature.this`) and apply again.
