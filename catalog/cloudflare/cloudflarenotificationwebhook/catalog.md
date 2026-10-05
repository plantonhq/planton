# Cloudflare Notification Webhook

Registers a webhook destination for Cloudflare alerting: the HTTPS endpoint — a Slack, Google Chat, Discord, Feishu, Datadog, Opsgenie, or Splunk integration, or any generic receiver you run — that Cloudflare Notification Policy resources deliver alerts to. Cloudflare inspects the URL, infers the destination type itself, and shapes each alert payload accordingly; the inferred type is reported back as an output. A plain CRUD resource: real create, update, and delete, with only the account forcing replacement.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Webhook Destination** -- one `cloudflare_notification_policy_webhooks` in the account's destinations list, with the optional shared secret attached when one is provided

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Notifications Write on the account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **A live HTTPS endpoint** -- a vendor incoming-webhook URL (Slack, Google Chat, Discord) or your own receiver. Cloudflare accepts an unreachable or wrong URL at registration; delivery is only proven when an alert actually fires.
- **The vendor API key** (only for Datadog, Splunk, and Opsgenie destinations) -- it travels in the `secret` field, which is why that field is sensitive regardless of destination kind. Keep it in a managed secret and reference it.

## Deploy

### Console

Open the deployment store, find **Cloudflare Notification Webhook**, and click **Deploy**. The creation wizard walks you through the account, the destination name, the endpoint URL, and the optional shared secret. Start from the **Slack channel** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareNotificationWebhook
metadata:
  name: ops-slack
  org: acme-corp
  env: prod
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: ops-slack
  url: "https://hooks.slack.com/services/your-workspace/your-incoming-webhook"
```

```shell
planton apply -f notification-webhook.yaml
```

This registers a Slack incoming-webhook URL as an alert destination — Cloudflare detects `slack` from the URL and formats alert messages for it. Reference the resulting `webhook_id` from any notification policy. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a notification webhook. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The type is a server-side echo, not a choice.** Cloudflare classifies the destination from the URL — Slack, Google Chat, Discord, Feishu, Datadog, Opsgenie, Splunk, or generic — and the classification changes how the payload is shaped. There is no override: if you need Slack-formatted messages, the URL must be a Slack incoming-webhook URL; a generic proxy in front of Slack gets generic payloads.

**The secret is write-only, forever.** Cloudflare never returns `secret` in any API response — drift on it cannot be detected, an imported destination arrives with an empty secret, and re-applying is what puts your configured value back. Provide a managed-secret reference so the platform resolves it just-in-time at deploy, and treat rotation as "set the new value and apply," never as "check what's there."

**For Datadog, Splunk, and Opsgenie, `secret` carries the vendor API key.** It is not a verify-the-sender shared secret in those cases — the same field does double duty by Cloudflare's design.

**Deleting a destination silently breaks its policies.** Policies reference destinations by UUID; delete the destination and every policy that used it simply stops delivering on that channel — no error, no plan diff in the policies, no symptom until an alert you counted on does not arrive. Retire the policy references first, then the destination; when several policies share one destination, that ordering matters more than it looks.

**Registration does not prove delivery.** After registering, cause one alert deliberately (a health check you can fail) and confirm it landed — the destination's last-success and last-failure timestamps only populate once an alert actually fires. Once per channel is enough to prove the pipe.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. The endpoint URL is a literal, and the shared secret arrives as a managed-secret reference rather than a typed component reference.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `webhook_id` | The Cloudflare-assigned UUID of the destination | Referenced by `mechanisms.webhookIds` on Cloudflare Notification Policy resources |
| `type` | The destination type Cloudflare inferred from the URL (datadog, discord, feishu, gchat, generic, opsgenie, slack, splunk) | Verifying the URL classified as intended before policies depend on it |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Slack channel** -- A Slack incoming-webhook URL registered with no secret; the URL itself is the credential Slack issues. Cloudflare detects `slack` and formats messages for the channel. Start from the **Slack channel** preset.

**Generic receiver with a shared secret** -- Alerts POSTed to a service you run, with a shared secret your receiver checks so it can reject anything that did not come from Cloudflare. Cloudflare classifies it as `generic` and sends the standard payload. Start from the **Generic receiver with a shared secret** preset.

**One destination, many policies** -- Register the channel once and point every relevant notification policy's `webhookIds` at the same `webhook_id` via ValueFromRef — which is also why the delete ordering above deserves respect.

## Works With

- [**Cloudflare Notification Policy**](/infra-catalog/cloudflare-notification-policy) -- the alert rules that deliver to this destination; they consume `webhook_id` via ValueFromRef.
- [**Cloudflare Logpush Job**](/infra-catalog/cloudflare-logpush-job) -- the record-level counterpart: continuous log delivery where this kind carries discrete alert events.
