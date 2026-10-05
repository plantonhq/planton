# GCP Monitoring Notification Channel

Creates a Cloud Monitoring notification channel — the delivery endpoint (email, Slack, PagerDuty, SMS, webhook, or Pub/Sub) that alerting policies notify when incidents open, close, or gain new violations. Credentials for external services are handled as managed secrets end to end, and alert policies wire to the channel through ValueFromRef — the paging path becomes one declarative graph.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Notification Channel** -- a `monitoring.NotificationChannel` resource of the configured `type`, carrying its type-specific configuration and (for authenticated types) its credentials
- **Monitoring API enablement** -- `monitoring.googleapis.com` enabled in the target project (never disabled on destroy)
- **GCP Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically as `user_labels` for tracking and governance

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the channel is created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **IAM**: see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying identity needs.

### External Service (authenticated types)

- **Slack**: an OAuth token from the Slack app installation (`sensitiveLabels.authToken`).
- **PagerDuty**: the service integration key (`sensitiveLabels.serviceKey`).
- **Webhook with basic auth**: the endpoint password (`sensitiveLabels.password`).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpMonitoringNotificationChannel
metadata:
  name: oncall-email
spec:
  type: email
  channelLabels:
    email_address: oncall@example.com
```

```shell
planton apply -f channel.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `type` | `string` | Delivery mechanism: `email`, `sms`, `slack`, `pagerduty`, `webhook_tokenauth`, `webhook_basicauth`, `pubsub`, ... GCP validates against its live channel-type catalog. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default | GCP project. Can reference a GcpProject resource. |
| `displayName` | `string` | `metadata.name` | Name shown in the console and notification footers (≤512 chars). |
| `description` | `string` | `""` | Who owns the endpoint and why it exists (≤1024 bytes). |
| `channelLabels` | `map<string,string>` | `{}` | Type-specific NON-SECRET configuration (e.g. `email_address`, `channel_name`, `url`). Credential keys are refused here. |
| `sensitiveLabels` | `message` | — | Credentials: `authToken` (slack), `password` (webhook_basicauth), `serviceKey` (pagerduty). Managed secrets — never plaintext at rest. |
| `enabled` | `bool` | `true` | Whether notifications are forwarded. Disabled channels keep configuration and references but deliver nothing. |
| `forceDelete` | `bool` | `false` | If true, deletion proceeds even when alert policies still reference the channel (they lose it in the same operation). |
| `labels` | `map<string,string>` | `{}` | User metadata labels (maps to `user_labels`), merged with platform labels. |
| `deletionPolicy` | `string` | `DELETE` | What destroy does: `DELETE`, `PREVENT` (refuse), or `ABANDON` (keep delivering, drop from management). |

### Validation Rules

- **Credentials never in channelLabels**: `auth_token`, `password`, and `service_key` are refused in the plain config map — they belong in `sensitiveLabels`.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `channel_name` | `string` | `projects/{project}/notificationChannels/{id}` — the value alert policies reference |
| `verification_status` | `string` | Verification state; SMS/email channels deliver only after verification (types needing none report `VERIFICATION_STATUS_UNSPECIFIED`) |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **A channel sends nothing on its own** — it is pure configuration until a GcpMonitoringAlertPolicy references it.
- **SMS and email channels require verification** before they deliver; check `verification_status` after the first deploy.
- **`forceDelete: false` (default) fails deletion while policies still reference the channel** — the safe posture; a dangling policy silently loses its delivery endpoint otherwise.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- [GcpMonitoringAlertPolicy](/docs/catalog/gcp/gcpmonitoringalertpolicy) — the policy that notifies this channel when incidents open
- [GcpMonitoringUptimeCheck](/docs/catalog/gcp/gcpmonitoringuptimecheck) — the probe whose failures typically drive those policies
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project and API enablement

## Additional Resources

- [Notification Channels Overview](https://cloud.google.com/monitoring/support/notification-options)
- [NotificationChannel API Reference](https://cloud.google.com/monitoring/api/ref_v3/rest/v3/projects.notificationChannels)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
