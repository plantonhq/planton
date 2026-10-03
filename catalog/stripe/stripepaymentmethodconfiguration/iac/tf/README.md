# StripePaymentMethodConfiguration — OpenTofu Module

OpenTofu module that declares which payment methods checkout offers. Stripe kinds run on OpenTofu only; there is no Pulumi module.

## What It Creates

- `stripe_payment_method_configuration` — a new payment-method configuration in the account the provider's key belongs to (or, with `STRIPE_ACCOUNT`, a Connect account), with one display preference (`on`, `off`, `none`) per method the spec sets. It never adopts the account's default configuration; the application passes this configuration's `id` when it creates a Checkout Session or PaymentIntent. Destroy deactivates it (`active = false`) and Stripe keeps it forever. `parent` (Connect) is create-only and replaces the configuration when changed.

## Prerequisites

- [OpenTofu](https://opentofu.org/) (the provider is pinned exactly at `stripe/stripe` `0.3.0` from OpenTofu's registry).
- A Stripe key in `STRIPE_API_KEY`: a restricted key with "Payment Method Configurations" write. `STRIPE_ACCOUNT` optionally names a Connect account to act on.
- A method is offered only where its capability is active on the account; turn capabilities on in the Dashboard.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Catalog object metadata (`name`, `org`, `env`, ...) |
| `spec` | `name`, `active` (default `true`), `parent` (replaces); one `{ preference }` per payment method (`card`, `apple_pay`, `google_pay`, `link`, `sepa_debit`, ... 59 in all) |

## Outputs

| Name | Description |
|---|---|
| `id` | The configuration's Stripe id (`pmc_...`) |
| `is_default` | Whether it is the account's default configuration |
| `active` | Whether payments may use it |
| `available_payment_methods` | The methods Stripe reports available: set on and with their capability active |

## What the Module Does Not Do

A method removed from the manifest keeps its last preference in Stripe: the provider sends only values that are set. Set the method to `none` to hand it back to Stripe's default. French meal vouchers (`fr_meal_voucher_conecs`), which Stripe's API accepts, are not exposed by the pinned provider.
