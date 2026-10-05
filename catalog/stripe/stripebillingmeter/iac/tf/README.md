# StripeBillingMeter — OpenTofu Module

OpenTofu module that declares one Stripe billing meter and its usage alerts. Stripe kinds run on OpenTofu only; there is no Pulumi module.

## What It Creates

- `stripe_billing_meter` -- the meter, with `customer_mapping.type` set to `by_id`. Only `display_name` updates in place; any other change replaces it. Destroy deactivates it.
- `stripe_billing_alert` -- one alert per declared alert, keyed by its title. An alert cannot be updated (a change replaces it), and the provider's delete only forgets it: a replaced, removed or destroyed alert stays active in Stripe.

## Prerequisites

- [OpenTofu](https://opentofu.org/) (the provider is pinned exactly at `stripe/stripe` `0.3.0` from OpenTofu's registry).
- A Stripe key in `STRIPE_API_KEY`: a restricted key with "Billing Meters" write, and write on alerts when alerts are declared.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Catalog object metadata (`name`, `org`, `env`, ...) |
| `spec` | `display_name`, `event_name`, `default_aggregation` (required); `customer_mapping`, `value_settings`, `event_time_window`, `alerts` |

## Outputs

| Name | Description |
|---|---|
| `id` | The meter's Stripe id (`mtr_...`) |
| `event_name` | The name the application sends with each usage event |
| `status` | `active`, or `inactive` once deactivated |
| `alert_ids` | Each alert's title mapped to its Stripe id |

## Recovering a Meter That Cannot Be Read

The provider does not treat a missing meter as gone, so the next plan fails reading it. Remove it from state (`tofu state rm stripe_billing_meter.this`) and apply again.
