# StripePaymentMethodDomain — OpenTofu Module

OpenTofu module that registers one domain for wallet buttons. Stripe kinds run on OpenTofu only; there is no Pulumi module.

## What It Creates

- `stripe_payment_method_domain` -- the registration. `domain_name` is create-only, so changing it replaces the registration. Destroy only removes it from state: the domain stays registered, and enabled, in Stripe.

## Prerequisites

- [OpenTofu](https://opentofu.org/) (the provider is pinned exactly at `stripe/stripe` `0.3.0` from OpenTofu's registry).
- A Stripe key in `STRIPE_API_KEY`: a restricted key with "Payment Method Domains" write.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Catalog object metadata (`name`, `org`, `env`, ...) |
| `spec` | `domain_name` (required, replaces); `enabled` (optional, default true) |

## Outputs

| Name | Description |
|---|---|
| `id` | The registration's Stripe id (`pmd_...`) |
| `enabled` | Whether wallets may appear |
| `<wallet>_status`, `<wallet>_error_message` | Per wallet: `apple_pay`, `google_pay`, `link`, `paypal`, `amazon_pay`, `klarna` |

## Recovering a Domain That Cannot Be Read

The provider does not treat a missing domain as gone, so the next plan fails reading it. Remove it from state (`tofu state rm stripe_payment_method_domain.this`) and apply again.
