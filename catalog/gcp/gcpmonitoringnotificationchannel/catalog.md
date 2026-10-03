# GCP Monitoring Notification Channel

Creates a Cloud Monitoring notification channel — the delivery endpoint (email, Slack, PagerDuty, SMS, webhook, or Pub/Sub) that alerting policies notify when incidents open, close, or gain new violations. Credentials for external services are handled as managed secrets end to end, and alert policies wire to the channel through ValueFromRef — the paging path becomes one declarative graph.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Notification Channel** -- a `monitoring.NotificationChannel` of the configured `type` with its type-specific configuration and credentials
- **Monitoring API enablement** -- `monitoring.googleapis.com` enabled in the target project (never disabled on destroy)
- **GCP Labels** -- resource metadata labels applied automatically as `user_labels` for tracking and governance

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the channel is created (directly or via a GcpProject reference).
- **IAM**: the deploying identity needs `roles/monitoring.notificationChannelEditor` or broader.
- **External credentials** for authenticated types (a Slack OAuth token, a PagerDuty service key, a webhook password) — supplied as managed secrets through `sensitiveLabels`.

## Deploy

### Console

Open the deployment store, find **GCP Monitoring Notification Channel**, and click **Deploy**. The creation wizard walks you through the target project, the channel type, its configuration labels, and the credential fields for authenticated types. Start from the **On-call Email** preset in the [Presets](#presets) tab for the most common shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpMonitoringNotificationChannel
metadata:
  name: oncall-email
  org: acme-corp
  env: prod
spec:
  type: email
  channelLabels:
    email_address: oncall@example.com
```

```shell
planton apply -f channel.yaml
```

This creates an email channel that alert policies can reference. Email channels require verification before they deliver — check the `verification_status` output. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the channel references its project via ValueFromRef:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: observability-project
      fieldPath: status.outputs.project_id
  type: email
  channelLabels:
    email_address: oncall@example.com
```

The InfraPipeline deploys the project first, then the channel — and downstream alert policies in the same chart consume this channel's `channel_name` output in their `notificationChannels` list.

## Key Configuration

These are the most important decisions when configuring a notification channel. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Type** -- the delivery mechanism (`email`, `sms`, `slack`, `pagerduty`, `webhook_tokenauth`, `webhook_basicauth`, `pubsub`, ...). GCP validates the value and its configuration keys server-side against its live channel-type catalog.

**Channel labels vs sensitive labels** -- non-secret configuration (an email address, a Slack channel name, a webhook URL) goes in `channelLabels`; credentials (`authToken`, `password`, `serviceKey`) go in `sensitiveLabels`, where the platform enforces managed-secret handling. Credential keys in the plain map are refused by validation.

**Enablement** -- `enabled: false` silences the endpoint without rewiring policies; the channel keeps its configuration and references.

**Force delete** -- off by default, so deleting a channel that policies still reference FAILS instead of silently un-paging them.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `channel_name` | `projects/{project}/notificationChannels/{id}` | GcpMonitoringAlertPolicy `notificationChannels` — the paging composition key |
| `verification_status` | Verification state | Confirming SMS/email channels actually deliver |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**On-call email** -- the simplest delivery path; no external credentials, no verification beyond the address's inbox. Start from the **On-call Email** preset.

**Slack alerts channel** -- posts incidents into a Slack channel using the app installation's OAuth token (a managed secret). Start from the **Slack Channel** preset.

**PagerDuty escalation** -- routes incidents into a PagerDuty service via its integration key. Start from the **PagerDuty Service** preset.

## Works With

- [**GCP Monitoring Alert Policy**](/infra-catalog/gcp-monitoring-alert-policy) -- the rule that notifies this channel when incidents open
- [**GCP Monitoring Uptime Check**](/infra-catalog/gcp-monitoring-uptime-check) -- the probe whose failures typically drive those policies
- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the channel is created
