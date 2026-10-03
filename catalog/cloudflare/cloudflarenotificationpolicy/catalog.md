# Cloudflare Notification Policy

Deploys a Cloudflare notification policy: one alert type — a tunnel degrading, an origin failing health checks, a certificate approaching expiry, a DDoS event — fanned out to email addresses, PagerDuty services, and webhook destinations, optionally narrowed by filters. This spec requires at least one destination even though Cloudflare's API does not, because a destination-less policy is one Cloudflare happily accepts and never delivers from. A plain CRUD resource: real create, update, and delete, with only the account forcing replacement.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Notification Policy** -- one `cloudflare_notification_policy` watching the declared `alertType`, with each flattened destination (an address or a UUID) rebuilt into the API's object rows and only the declared filters sent

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Notifications Write on the account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **A plan carrying the chosen alert type** -- advanced DDoS (L4 and L7), bot traffic, and several script-monitor and security-insight families are Business or Enterprise features. Cloudflare refuses the create when the plan lacks the type — an honest failure, but at apply time, not review time.
- **A connected PagerDuty integration** (only for `pagerdutyIds`) -- the service UUID must come from an integration already connected in the Cloudflare dashboard. An unknown UUID is accepted at deploy and fails at delivery.
- **Live webhook destinations** (only for `webhookIds`) -- each UUID must point at an existing notification webhook, ideally a Cloudflare Notification Webhook Infra Component wired by reference.

## Deploy

### Console

Open the deployment store, find **Cloudflare Notification Policy**, and click **Deploy**. The creation wizard walks you through the account, the alert type, the destination mechanisms, and the per-type filters. Start from the **On-call paging** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareNotificationPolicy
metadata:
  name: tunnel-down-page-oncall
  org: acme-corp
  env: prod
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: tunnel-down-page-oncall
  alertType: tunnel_health_event
  description: page on-call the moment a Cloudflare Tunnel goes unhealthy
  alertInterval: 30m
  mechanisms:
    emails:
      - oncall@acme.com
  filters:
    newStatus:
      - unhealthy
```

```shell
planton apply -f notification-policy.yaml
```

This creates an enabled policy that emails on-call when any tunnel in the account turns unhealthy, refiring at most every 30 minutes so a flapping tunnel does not become an inbox storm. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the webhook destination to a notification webhook managed in the same InfraPipeline:

```yaml
spec:
  mechanisms:
    webhookIds:
      - valueFrom:
          kind: CloudflareNotificationWebhook
          name: platform-chat
          fieldPath: status.outputs.webhook_id
```

The InfraPipeline resolves the dependency graph, registers the webhook first, then provisions the policy pointing at its UUID.

## Key Configuration

These are the most important decisions when configuring a notification policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One policy, one alert type.** `alertType` is the event class the policy watches — the provider's list holds 69 values, from `tunnel_health_event` to `universal_ssl_event_type` to `failing_logpush_job_disabled_alert`. Cover each concern with its own policy rather than overloading one; deleting or editing a policy then has a blast radius of exactly one alert.

**Filters belong to alert types, and the pairing is API-owned.** The `filters` block carries 43 list fields because that is the provider's full set, but a given alert type reads only a handful — `incidentImpact` means something to `incident_alert` and nothing to `tunnel_health_event`. A filter the type ignores is accepted and silently does nothing, which is the quiet way to build a policy that never fires the way you expected. Every filter value travels as a string, even numeric thresholds and booleans — that is Cloudflare's grammar.

**Destinations must be real, not just present.** The spec forces at least one of `emails`, `pagerdutyIds`, or `webhookIds`, but being declared is not the same as working: neither PagerDuty UUIDs nor webhook UUIDs are validated at deploy, and both fail at delivery time — exactly when you are relying on the alert. Prove each channel once with an alert you can cause on purpose (a health check you can fail, a tunnel you can stop).

**The refire floor tames flapping.** `alertInterval` (e.g. `30m`, `2h`) sets the minimum time between successive notifications for the same condition. Only some alert types honor it; empty uses Cloudflare's per-type behavior.

**Deleting a policy is silent.** No destination is touched and no other policy notices — the only symptom is alerts that stop arriving, which nobody observes until the thing you were watching for happens. That is precisely why alert coverage belongs in manifests: it becomes reviewable infrastructure.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareNotificationWebhook** | `mechanisms.webhookIds[]` | `status.outputs.webhook_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `policy_id` | The Cloudflare-assigned UUID of the policy | API calls against the policy, cross-referencing delivery history in the dashboard |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**On-call paging** -- A tunnel turning unhealthy pages PagerDuty and copies the on-call inbox, with a 30-minute refire floor. Swap `tunnel_health_event` for `load_balancing_health_alert` or `http_alert_origin_error` to page on those instead. Start from the **On-call paging** preset.

**Certificate expiry warning** -- Universal SSL certificate events emailed to the platform team and mirrored into chat through a webhook destination. Cloudflare splits certificate expiry across five alert families (`universal_ssl_event_type`, `custom_ssl_certificate_event_type`, `dedicated_ssl_certificate_event_type`, `mtls_certificate_store_certificate_expiration_type`, and the Authenticated Origin Pulls variants) — clone the shape per certificate kind the account actually uses. Start from the **Certificate expiry warning** preset.

**Guarding log delivery** -- A policy on `failing_logpush_job_disabled_alert` beside every Logpush job that matters for audit or compliance, closing the gap where log delivery stops and nothing else notices.

## Works With

- [**Cloudflare Notification Webhook**](/infra-catalog/cloudflare-notification-webhook) -- registers the webhook destinations this policy delivers to; wire `webhookIds` via ValueFromRef.
- [**Cloudflare Health Check**](/infra-catalog/cloudflare-healthcheck) -- the origin probes behind `health_check_status_notification` policies.
- [**Cloudflare Logpush Job**](/infra-catalog/cloudflare-logpush-job) -- pair a policy on the failing-job alert with every job whose delivery must not stop silently.
