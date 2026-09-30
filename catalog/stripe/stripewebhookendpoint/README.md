# StripeWebhookEndpoint

Declares where a Stripe account delivers its [events](https://docs.stripe.com/webhooks): the URL, the event types, and the signing secret the receiving service verifies every delivery with.

## When to Use

- **Payment events reach your application**: a completed checkout, a paid invoice, a refund arrive at your webhook route the moment they happen.
- **No secret copied by hand**: the signing secret is captured when the endpoint is created and read by reference by the service that verifies deliveries.
- **The same endpoint in every environment**: each account's endpoint is a file, so a new environment gets the same events without a Dashboard visit.

## Quick Start

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeWebhookEndpoint
metadata:
  name: billing-events
  org: acme-corp
  env: production
spec:
  url: https://app.acme.com/stripe
  enabledEvents:
    - checkout.session.completed
    - invoice.paid
    - charge.refunded
```

## Fields

| Field | Description |
|---|---|
| `url` | Where Stripe POSTs each event (required). Live mode requires HTTPS; a sandbox accepts HTTP. Changes in place |
| `enabledEvents` | The event types delivered (required), in Stripe's spelling; `*` for every event except those that must be named |
| `description` | A note shown in the Dashboard |
| `metadata` | Key-value pairs stored on the endpoint |
| `apiVersion` | Pins the API version events render in. Changing it **replaces** the endpoint and rotates its secret |
| `connect` | Delivers events from the account's connected accounts. Changing it **replaces** the endpoint and rotates its secret |

## Key Behaviors

- **The secret exists only at creation.** It lands in `status.outputs.secret`. An endpoint imported from one made in the Dashboard has no secret; replace it to get one.
- **Destroy deletes the endpoint**, and its secret with it.
- **Deleted outside Planton, the next plan fails**: the provider does not treat a missing endpoint as gone. Remove it from state and apply again for a new one.
- **Runs on OpenTofu only.** Planton refuses any other engine for this kind before anything runs.

## Outputs

| Output | Description |
|---|---|
| `id` | The endpoint's Stripe id (`we_...`) |
| `secret` | The signing secret (`whsec_...`), sensitive |
| `status` | `enabled` or `disabled`, as Stripe reports it |
| `url` | The address events are delivered to |
| `application` | The Connect application that created the endpoint, when one did |
