# Stripe Webhook Endpoint

Declares where a Stripe account delivers its events -- the URL and the event types -- and captures the signing secret your service verifies every delivery with. One Infra Component per endpoint.

## What Gets Created

When you deploy this Infra Component, the OpenTofu module creates one webhook endpoint in the Stripe account your Stripe connection's key belongs to:

- **The delivery address** -- the URL Stripe POSTs each event to
- **The events** -- exactly the event types you list
- **The signing secret** -- returned by Stripe only now, and kept in `status.outputs.secret`

## Before You Deploy

### Planton Setup

- **Stripe Provider Connection** -- an active connection with a restricted key and its mode (test or live). A key of the other mode is refused before anything runs.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Stripe Account

- **The connection's restricted key** needs **Webhook Endpoints: Write** (Stripe Dashboard: Developers, API keys, the key's permissions).
- **A route that answers**: Stripe does not check the URL when the endpoint is created, but disables an endpoint whose deliveries keep failing.

## Deploy

### Console

Open the deployment store, find **Stripe Webhook Endpoint**, and click **Deploy**. Start from the **Billing Events** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeWebhookEndpoint
metadata:
  name: billing-events
  org: acme-corp
  env: prod
spec:
  url: https://app.acme.com/stripe
  enabledEvents:
    - checkout.session.completed
    - invoice.paid
    - charge.refunded
```

```shell
planton apply -f stripe-webhook-endpoint.yaml
```

The endpoint starts receiving events at once. An Infra Job tracks the change in real time.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Only the events you handle** -- `enabledEvents` lists them; every other delivery is traffic your route must acknowledge and ignore.

**Verify every delivery** -- your service checks each `Stripe-Signature` header with `status.outputs.secret`. Reference the output; never copy it.

**Leave `apiVersion` unset** -- changing it replaces the endpoint, and a new endpoint has a new secret.

**Destroy deletes** -- the endpoint and its secret are gone; recreating it means a new secret.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies.

### What This Kind Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `id` | The endpoint's Stripe id | Audits, support |
| `secret` | The signing secret (sensitive) | The receiving service's webhook verification, by reference |
| `status` | `enabled` or `disabled` | Monitoring |
| `url` | The delivery address | Audits |
| `application` | The creating Connect application | Connect platforms |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Billing events** -- Checkout completions, paid invoices and refunds for one application route. Start from the **Billing Events** preset.

**Subscription lifecycle** -- Every subscription change for a billing service that mirrors subscription state. Start from the **Subscription Lifecycle** preset.

## Works With

- [**Stripe Billing Portal Configuration**](/infra-catalog/stripe-billing-portal-configuration) -- the portal whose cancellations and plan changes arrive here as subscription events.
- [**Stripe Payment Method Configuration**](/infra-catalog/stripe-payment-method-configuration) -- the methods checkout offers; the checkouts they complete arrive here.
