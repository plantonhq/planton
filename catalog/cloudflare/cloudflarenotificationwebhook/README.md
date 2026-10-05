# Cloudflare Notification Webhook

## Overview

`CloudflareNotificationWebhook` registers a webhook destination for Cloudflare's alerting system: an HTTPS endpoint that `CloudflareNotificationPolicy` resources deliver alerts to. Cloudflare recognises the popular targets by their URL shape -- Slack, Google Chat, Discord, Feishu, Datadog, Opsgenie, Splunk -- and treats anything else as a generic receiver. A plain CRUD object -- real create, update, delete.

## Key Features

- **Any HTTPS receiver** -- vendor integrations by their incoming-webhook or intake URL, or your own service
- **Server-inferred type** -- Cloudflare reports the destination kind it detected (one of eight values) as an output; it is never configured
- **Shared-secret authentication** -- an optional write-only secret the receiver can use to verify deliveries came from Cloudflare (and the API-key slot for Datadog, Splunk, and Opsgenie)
- **Delivery observability** -- Cloudflare records last-success and last-failure timestamps on the destination

## Use Cases

**Ideal for:**

- Mirroring Cloudflare alerts into a Slack or Google Chat channel
- Forwarding alerts into Datadog, Splunk, or Opsgenie as events
- Fanning alerts into your own automation service

**Not ideal for:**

- The alert rules themselves -- those are `CloudflareNotificationPolicy` resources that reference this destination
- Log delivery -- webhooks carry alert events, not log records; use `CloudflareLogpushJob`

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `account_id` | string | Yes | The Cloudflare account (32-hex). |
| `name` | string | Yes | Shown in the dashboard's destinations list. |
| `url` | string | Yes | The endpoint alerts are POSTed to. |

### Key Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `secret` | string reference | Shared secret sent with each delivery (or the vendor API key for Datadog, Splunk, Opsgenie). Sensitive and write-only -- Cloudflare never returns it. |

### Outputs

| Field | Description |
|-------|-------------|
| `webhook_id` | The Cloudflare-assigned UUID (what notification policies reference) |
| `type` | The destination kind Cloudflare inferred from the URL |

## Example Manifest

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareNotificationWebhook
metadata:
  name: ops-slack
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: ops-slack
  url: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXX"
  secret:
    valueFrom:
      kind: SecretsManagerSecret
      name: slack-webhook-secret
      fieldPath: status.outputs.secret_value
```

## Prerequisites

- **A Cloudflare API token** with the Notifications Write permission
- **A live HTTPS endpoint** -- the vendor integration's webhook URL, or your own receiver

## Destroy Semantics

Real delete. Deleting the destination drops it from every notification policy that referenced it, and those policies simply stop delivering on that channel -- no error, no warning.

## Related Kinds

- [Cloudflare Notification Policy](/docs/catalog/cloudflare/cloudflarenotificationpolicy) -- the alert rules that deliver here
- [Cloudflare Logpush Job](/docs/catalog/cloudflare/cloudflarelogpushjob) -- log delivery, the record-level counterpart to alerting

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
