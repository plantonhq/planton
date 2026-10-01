# StripeWebhookEndpoint — OpenTofu Module

OpenTofu module that declares where a Stripe account delivers its events. Stripe kinds run on OpenTofu only; there is no Pulumi module.

## What It Creates

- `stripe_webhook_endpoint` — an endpoint in the account the provider's key belongs to (or, with `STRIPE_ACCOUNT`, a Connect account), delivering the listed event types to one URL. Its signing secret is captured at creation into the `secret` output. `api_version` and `connect` are create-only, so changing either replaces the endpoint and rotates the secret. Destroy deletes the endpoint.

## Prerequisites

- [OpenTofu](https://opentofu.org/) (the provider is pinned exactly at `stripe/stripe` `0.3.0` from OpenTofu's registry).
- A Stripe key in `STRIPE_API_KEY`: a restricted key with "Webhook Endpoints" write. `STRIPE_ACCOUNT` optionally names a Connect account to act on.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Cloud resource metadata (`name`, `org`, `env`, ...) |
| `spec` | `url`, `enabled_events` (required); `description`, `metadata`, `api_version` (replaces), `connect` (replaces) (optional) |

## Outputs

| Name | Description |
|---|---|
| `id` | The endpoint's Stripe id (`we_...`) |
| `secret` | The signing secret (sensitive); empty for an imported endpoint |
| `status` | `enabled` or `disabled`, as Stripe reports it |
| `url` | The address events are delivered to |
| `application` | The Connect application that created the endpoint, when one did |

## Recovering an Endpoint Deleted Outside Planton

The provider does not treat a missing endpoint as gone, so the next plan fails reading it. Remove it from state (`tofu state rm stripe_webhook_endpoint.this`) and apply again: a new endpoint is created, with a new signing secret.
