# StripePaymentLink — OpenTofu Module

OpenTofu module that declares one Stripe payment link. Stripe kinds run on OpenTofu only; there is no Pulumi module.

## What It Creates

- `stripe_payment_link` -- the link, with `create_before_destroy`, so a replacement creates the new link (and its new address) before the old one is deactivated. Destroy deactivates it. `line_items[*].price_data` is never set.
- `terraform_data.line_items` -- a tracker of each line item's price and quantity, in order. Stripe can't change a line item's price, and the provider never stores its quantity, so a change to either would otherwise fail or be dropped; a change to the tracker replaces the link. The tracker holds nothing in Stripe and is never imported.

## Prerequisites

- [OpenTofu](https://opentofu.org/) (the provider is pinned exactly at `stripe/stripe` `0.3.0` from OpenTofu's registry).
- A Stripe key in `STRIPE_API_KEY`: a restricted key with "Payment Links" write.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Cloud resource metadata (`name`, `org`, `env`, ...) |
| `spec` | `line_items` (resolved price ids, 1 to 20, required); `optional_items`, `active`, `inactive_message`, `after_completion`, `allow_promotion_codes`, `automatic_tax`, `billing_address_collection`, `consent_collection`, `currency`, `custom_fields`, `custom_text`, `customer_creation`, `invoice_creation`, `managed_payments`, `name_collection`, `payment_intent_data`, `payment_method_collection`, `payment_method_options`, `payment_method_types`, `phone_number_collection`, `restrictions`, `shipping_address_collection`, `shipping_options` (resolved shipping rate ids), `submit_type`, `subscription_data`, `tax_id_collection`, `application_fee_amount`, `application_fee_percent`, `on_behalf_of`, `transfer_data`, `metadata` |

## Outputs

| Name | Description |
|---|---|
| `id` | The payment link's Stripe id (`plink_...`) |
| `url` | The page's public address |
| `active` | `false` once deactivated |

## Recovering a Payment Link That Cannot Be Read

The provider does not treat a missing payment link as gone, so the next plan fails reading it. Remove it from state (`tofu state rm stripe_payment_link.this`) and apply again.
