# StripeBillingPortalConfiguration — OpenTofu Module

OpenTofu module that declares what Stripe's customer portal lets a customer do. Stripe kinds run on OpenTofu only; there is no Pulumi module.

## What It Creates

- `stripe_billing_portal_configuration` — a new portal configuration in the account the provider's key belongs to (or, with `STRIPE_ACCOUNT`, a Connect account). It never adopts the account's default configuration; the application passes this configuration's `id` when it opens a portal session. Destroy deactivates it (`active = false`) and Stripe keeps it forever.

## Prerequisites

- [OpenTofu](https://opentofu.org/) (the provider is pinned exactly at `stripe/stripe` `0.3.0` from OpenTofu's registry).
- A Stripe key in `STRIPE_API_KEY`: a restricted key with "Customer portal" write. `STRIPE_ACCOUNT` optionally names a Connect account to act on.
- `subscription_update` is refused by the spec on the pinned provider: Stripe requires switchable products whenever it is on, returns them only when a read expands them, and the provider never does, so a create with them fails. The module still writes the feature, for a provider version that reads it back.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Cloud resource metadata (`name`, `org`, `env`, ...) |
| `spec` | `features` (required: `customer_update`, `invoice_history`, `payment_method_update`, `subscription_cancel`, `subscription_update`); `name`, `business_profile`, `default_return_url`, `login_page`, `active` (default `true`), `metadata` (optional) |

## Outputs

| Name | Description |
|---|---|
| `id` | The configuration's Stripe id (`bpc_...`) |
| `is_default` | Whether it is the account's default configuration |
| `active` | Whether portal sessions may use it |
| `login_page_url` | The shareable sign-in URL, when `login_page` is enabled |

## What the Module Does Not Do

A setting removed from the manifest is not reset in Stripe: the provider sends only values that are set. Change a setting to its opposite instead (for example `enabled: false`).
