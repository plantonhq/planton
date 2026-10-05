# StripeRadarValueList — OpenTofu Module

OpenTofu module that declares one Stripe Radar list and its items. Stripe kinds run on OpenTofu only; there is no Pulumi module.

## What It Creates

- `stripe_radar_value_list` -- the list. `item_type` is create-only, so changing it replaces the list. Destroy deletes it.
- `stripe_radar_value_list_item` -- one item per value, keyed by the value. An item cannot be updated; removing a value deletes its item.

## Prerequisites

- [OpenTofu](https://opentofu.org/) (the provider is pinned exactly at `stripe/stripe` `0.3.0` from OpenTofu's registry).
- A Stripe key in `STRIPE_API_KEY`: a restricted key with "Radar" write.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Catalog object metadata (`name`, `org`, `env`, ...) |
| `spec` | `alias`, `name` (required); `item_type` (replaces), `items`, `metadata` (optional) |

## Outputs

| Name | Description |
|---|---|
| `id` | The list's Stripe id (`rsl_...`) |
| `alias` | The name rules reference |
| `item_ids` | Each item's value mapped to its id (`rsli_...`) |

## Recovering a List That Cannot Be Read

The provider does not treat a missing list as gone, so the next plan fails reading it. Remove it from state (`tofu state rm stripe_radar_value_list.this`, and each `stripe_radar_value_list_item.this[...]`) and apply again.
