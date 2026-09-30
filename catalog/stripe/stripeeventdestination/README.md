# StripeEventDestination

Declares where a Stripe account sends its events through Stripe's v2 [event destinations](https://docs.stripe.com/event-destinations): a webhook URL, an Amazon EventBridge event bus, or an Azure Event Grid partner topic, as thin or snapshot events.

## When to Use

- **Thin events**: small notifications for Stripe's v2 APIs, which your service fetches details for.
- **Events into a cloud event bus**: Amazon EventBridge or Azure Event Grid, with no webhook route to run.
- **Events from other accounts**: routed from accounts you manage or your organization's members.

For classic snapshot webhooks to one URL, StripeWebhookEndpoint is the simpler kind.

## Quick Start

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeEventDestination
metadata:
  name: billing-thin-events
  org: acme-corp
  env: production
spec:
  name: Billing thin events
  eventPayload: thin
  enabledEvents:
    - v1.billing.meter.error_report_triggered
  webhookEndpoint:
    url: https://app.acme.com/stripe/thin
```

## Fields

| Field | Description |
|---|---|
| `name` | The destination's name (required). Changes in place |
| `description` | What it is for. Changes in place |
| `eventPayload` | `thin` or `snapshot` (required). **Replaces** |
| `enabledEvents` | Event types delivered (required). Changes in place |
| `eventsFrom` | Whose events: `@self`, `@accounts`, `@organization_members`, `@organization_members/@accounts`. **Replaces** |
| `snapshotApiVersion` | The API version snapshot events render in. **Replaces** |
| `webhookEndpoint` / `amazonEventbridge` / `azureEventGrid` | Exactly one: where events go, and so the destination's type |
| `metadata` | Key-value pairs stored on the destination |

## Key Behaviors

- **A webhook's signing secret exists only at creation.** It lands in `status.outputs.signing_secret`; an imported destination has none, and a replacement yields a new one.
- **EventBridge and Event Grid wait for you.** The destination stays pending until the partner source is associated in AWS or activated in Azure; `status.outputs` names it.
- **Destroy deletes the destination.** The provider cannot enable or disable it, so `status` is observed only.
- **Runs on OpenTofu only.** Planton refuses any other engine for this kind before anything runs.

## Outputs

| Output | Description |
|---|---|
| `id` | The destination's Stripe id |
| `status`, `status_disabled_reason` | `enabled` or `disabled`, and why |
| `signing_secret` | A webhook destination's signing secret (sensitive) |
| `aws_event_source_arn`, `aws_event_source_name`, `aws_event_source_status` | The EventBridge partner source |
| `azure_partner_topic_name`, `azure_partner_topic_status` | The Event Grid partner topic |

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
