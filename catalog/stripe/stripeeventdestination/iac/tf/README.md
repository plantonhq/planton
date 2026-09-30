# StripeEventDestination — OpenTofu Module

OpenTofu module that declares one Stripe v2 event destination. Stripe kinds run on OpenTofu only; there is no Pulumi module.

## What It Creates

- `stripe_v2_core_event_destination` -- the destination. Its `type` is the destination block the spec sets. `event_payload`, `events_from`, `snapshot_api_version` and the EventBridge and Event Grid blocks are create-only, so changing any replaces the destination. A webhook destination's signing secret is captured at creation into the `signing_secret` output. Destroy deletes it.

## Prerequisites

- [OpenTofu](https://opentofu.org/) (the provider is pinned exactly at `stripe/stripe` `0.3.0` from OpenTofu's registry).
- A Stripe key in `STRIPE_API_KEY`: a restricted key with "Event Destinations" write.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Cloud resource metadata (`name`, `org`, `env`, ...) |
| `spec` | `name`, `event_payload`, `enabled_events` (required); exactly one of `webhook_endpoint`, `amazon_eventbridge`, `azure_event_grid`; `description`, `events_from`, `snapshot_api_version`, `metadata` (optional) |

## Outputs

| Name | Description |
|---|---|
| `id` | The destination's Stripe id |
| `status`, `status_disabled_reason` | As Stripe reports them |
| `signing_secret` | A webhook destination's signing secret (sensitive); empty after import |
| `aws_event_source_arn`, `aws_event_source_name`, `aws_event_source_status` | The EventBridge partner source |
| `azure_partner_topic_name`, `azure_partner_topic_status` | The Event Grid partner topic |

## Recovering a Destination That Cannot Be Read

The provider does not treat a missing destination as gone, so the next plan fails reading it. Remove it from state (`tofu state rm stripe_v2_core_event_destination.this`) and apply again: a new destination is created, and a webhook destination gets a new secret.
