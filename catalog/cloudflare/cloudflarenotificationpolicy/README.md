# Cloudflare Notification Policy

## Overview

`CloudflareNotificationPolicy` creates a notification policy: "when THIS happens, tell THESE people." One policy watches one alert type -- a tunnel going down, an origin failing health checks, a DDoS event, a usage threshold, a certificate nearing expiry -- and fans the alert out to email addresses, PagerDuty services, and webhook destinations. A plain CRUD object -- real create, update, delete.

## Key Features

- **69 alert types** -- the provider's full list at the pinned version, enum-walled: DDoS and traffic anomalies, origin and edge error rates, load-balancer and health-check transitions, tunnel health, certificate expiry across five certificate families, Pages deployments, Stream events, billing thresholds, script-monitor findings, secondary DNS, incidents, and more
- **Three delivery channels** -- email addresses, PagerDuty services, and webhook destinations, mixed freely; at least one is required
- **43 filter fields** -- narrow which events of the type actually fire, using the provider's own filter grammar (each field's comment names the alert families that read it)
- **Refire control** -- `alert_interval` bounds how often the same condition re-notifies, for the types that honor it

## Use Cases

**Ideal for:**

- Paging on-call through PagerDuty when a Cloudflare Tunnel or origin goes unhealthy
- Emailing the platform team when a certificate is about to expire
- Routing security alerts into an incident channel via a webhook destination

**Not ideal for:**

- The webhook endpoints themselves -- register those as `CloudflareNotificationWebhook` and reference them here
- Log export -- alerts are events, not records; use `CloudflareLogpushJob` for the log stream

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `account_id` | string | Yes | The Cloudflare account (32-hex). |
| `name` | string | Yes | Shown in the dashboard's notifications list. |
| `alert_type` | string | Yes | The event class watched (69 provider values, enum-walled). |
| `mechanisms` | object | Yes | Where alerts go; at least one destination is required. |

### Key Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `description` | string | Why the policy exists. |
| `enabled` | bool | Whether the policy fires (Cloudflare's default is enabled). |
| `alert_interval` | string | Minimum time between repeat notifications, e.g. `30m`. |
| `filters` | object | 43 list fields narrowing which events fire; `incident_impact` and `traffic_exclusions` are enum-walled. |

### Outputs

| Field | Description |
|-------|-------------|
| `policy_id` | The Cloudflare-assigned UUID of the policy |

## Example Manifest

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareNotificationPolicy
metadata:
  name: origin-health-alerts
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: origin-health-alerts
  alert_type: health_check_status_notification
  description: page on-call when an origin health check turns unhealthy
  mechanisms:
    emails:
      - oncall@example.com
    webhook_ids:
      - valueFrom:
          kind: CloudflareNotificationWebhook
          name: ops-slack
          fieldPath: status.outputs.webhook_id
  filters:
    status:
      - Unhealthy
```

## Prerequisites

- **A Cloudflare API token** with the Notifications Write permission
- **A plan carrying the chosen alert type** -- several families are Business or Enterprise only
- For PagerDuty delivery: the PagerDuty integration connected in the Cloudflare dashboard (its service UUID goes in `pagerduty_ids`)

## Destroy Semantics

Real delete. Deleting the policy stops the alerts it delivered, silently -- nothing else changes and no destination is affected.

## Related Kinds

- [Cloudflare Notification Webhook](/docs/catalog/cloudflare/cloudflarenotificationwebhook) -- the webhook destinations referenced here
- [Cloudflare Healthcheck](/docs/catalog/cloudflare/cloudflarehealthcheck) -- the origin probes behind health-check alerts
- [Cloudflare Logpush Job](/docs/catalog/cloudflare/cloudflarelogpushjob) -- log export, and a subject of the failing-job alert type

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
