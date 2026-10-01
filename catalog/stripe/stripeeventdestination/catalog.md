# Stripe Event Destination

Declares where your Stripe account sends events through Stripe's v2 event destinations -- a webhook URL, an Amazon EventBridge event bus, or an Azure Event Grid partner topic -- as thin or snapshot events. One Cloud Resource per destination.

## What Gets Created

When you deploy this Cloud Resource, the OpenTofu module creates one event destination in the Stripe account your Stripe connection's key belongs to:

- **The destination** -- a webhook URL, or a partner event source in your AWS account, or a partner topic in your Azure subscription
- **The events** -- exactly the event types you list, thin or snapshot
- **A webhook's signing secret** -- returned by Stripe only now, and kept in `status.outputs.signing_secret`

## Before You Deploy

### Planton Setup

- **Stripe Provider Connection** -- an active connection with a restricted key and its mode (test or live). A key of the other mode is refused before anything runs.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Stripe Account

- **The connection's restricted key** needs **Event Destinations: Write** (Stripe Dashboard: Developers, API keys, the key's permissions).
- **For EventBridge or Event Grid**: the AWS account or Azure subscription that will receive events.

## Deploy

### Console

Open the deployment store, find **Stripe Event Destination**, and click **Deploy**. Start from the **Thin Events Webhook** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeEventDestination
metadata:
  name: billing-thin-events
  org: acme-corp
  env: prod
spec:
  name: Billing thin events
  eventPayload: thin
  enabledEvents:
    - v1.billing.meter.error_report_triggered
  webhookEndpoint:
    url: https://app.acme.com/stripe/thin
```

```shell
planton apply -f stripe-event-destination.yaml
```

A Stack Job tracks the change in real time.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Thin or snapshot** -- thin events carry a notification your service fetches details for; snapshot events carry the whole object. Changing it replaces the destination.

**One destination block** -- `webhookEndpoint`, `amazonEventbridge` or `azureEventGrid`; the block you set is the destination's type.

**Verify every webhook delivery** -- your service checks each `Stripe-Signature` header with `status.outputs.signing_secret`, by reference.

**Destroy deletes** -- the destination, and a webhook's secret with it.

## Outputs and Dependencies

### What This Component Consumes

This component has no foreign key dependencies.

### What This Component Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `id` | The destination's Stripe id | Audits |
| `signing_secret` | A webhook's signing secret (sensitive) | The receiving service's verification, by reference |
| `aws_event_source_name` | The EventBridge partner source | An AWS EventBridge Bus's `eventSourceName` |
| `azure_partner_topic_name` | The Event Grid partner topic | Activation in Azure |
| `status` | `enabled` or `disabled` | Monitoring |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Thin events webhook** -- meter errors to a billing service. Start from the **Thin Events Webhook** preset.

**Amazon EventBridge** -- billing events into an AWS event bus. Start from the **Amazon EventBridge** preset.

## Works With

- [**Stripe Webhook Endpoint**](/cloud-catalog/stripe-webhook-endpoint) -- the simpler kind for classic snapshot webhooks.
- [**AWS EventBridge Bus**](/cloud-catalog/aws-event-bridge-bus) -- the bus that associates with an EventBridge destination's partner source.
