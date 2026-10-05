# GCP Monitoring Alert Policy

Creates a Cloud Monitoring alerting policy — the rule that watches metrics or logs (threshold, absence, log-match, MQL, PromQL, or scheduled-SQL conditions) and opens an incident, notifying the referenced channels, when its conditions are met. Channels wire in through ValueFromRef, so the whole paging path — check, policy, channel — is one declarative graph.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Alert Policy** -- a `monitoring.AlertPolicy` with the configured conditions, combiner, severity, notification channels, alert strategy, and runbook documentation
- **Monitoring API enablement** -- `monitoring.googleapis.com` enabled in the target project (never disabled on destroy)
- **GCP Labels** -- resource metadata labels applied automatically as `user_labels`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery.

### GCP Project

- **A GCP project** where the policy is created (directly or via a GcpProject reference).
- **IAM**: see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying identity needs.
- **Notification channels** to page (reference GcpMonitoringNotificationChannel resources).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpMonitoringAlertPolicy
metadata:
  name: cpu-saturation
spec:
  combiner: OR
  severity: WARNING
  conditions:
    - displayName: cpu above 80%
      conditionThreshold:
        filter: metric.type="compute.googleapis.com/instance/cpu/utilization" AND resource.type="gce_instance"
        comparison: COMPARISON_GT
        thresholdValue: 0.8
        duration: 300s
        aggregations:
          - alignmentPeriod: 60s
            perSeriesAligner: ALIGN_MEAN
  notificationChannels:
    - value: projects/my-project/notificationChannels/1234567890
```

```shell
planton apply -f alert-policy.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `combiner` | `string` | How condition results combine: `AND`, `OR`, `AND_WITH_MATCHING_RESOURCE`. Required by GCP even for single-condition policies. | Enum |
| `conditions` | `list` | 1–6 conditions, each with a display name and exactly one condition-type arm. | 1–6 items; exactly-one arm |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default | GCP project. Can reference a GcpProject resource. |
| `displayName` | `string` | `metadata.name` | Console and notification display name. |
| `severity` | `string` | unset | `CRITICAL`, `ERROR`, or `WARNING` on opened incidents. |
| `enabled` | `bool` | `true` | Whether the policy evaluates. Sent explicitly by both engines. |
| `notificationChannels` | `list<StringValueOrRef>` | `[]` | Channel resource names; reference GcpMonitoringNotificationChannel outputs. |
| `alertStrategy` | `message` | — | Auto-close, log-alert rate limiting, per-channel re-notification, prompts. |
| `documentation` | `message` | — | Runbook content (Markdown, `${variable}` substitution), subject, up to 3 links. |
| `labels` | `map<string,string>` | `{}` | User metadata labels, merged with platform labels. |
| `deletionPolicy` | `string` | `DELETE` | What destroy does: `DELETE`, `PREVENT`, `ABANDON`. |

Condition arms: `conditionThreshold` (the workhorse — filter, comparison, threshold, duration, aggregations, ratio denominators, forecasting, trigger, missing-data behavior), `conditionAbsent` (silence is failure), `conditionMatchedLog` (log-match; requires `alertStrategy.notificationRateLimit`), `conditionMonitoringQueryLanguage` (MQL — deprecated by Google; prefer PromQL), `conditionPrometheusQueryLanguage` (PromQL with duration/interval/labels), `conditionSql` (scheduled SQL over log analytics with a row-count or boolean test).

### Validation Rules

- **Exactly one condition arm per condition** — the API's oneof, enforced client-side here (the Terraform provider leaves it to the API).
- **SQL conditions** take exactly one schedule (minutes/hourly/daily) and exactly one test (rowCountTest/booleanTest).
- **Triggers** set `count` or `percent`, never both.
- **Documentation** `mimeType` accepts only `text/markdown`; at most 3 links.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `policy_name` | `string` | `projects/{project}/alertPolicies/{id}` |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md).

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md).

## Important Notes

- **Log-based policies** (`conditionMatchedLog`) REQUIRE `alertStrategy.notificationRateLimit` — the API's own pairing; metric policies reject it.
- **`thresholdValue: 0` is a real threshold** ("any errors at all") and is always sent.
- **A disabled policy keeps its configuration** but opens no incidents — the safe way to silence a noisy rule while tuning it.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- [GcpMonitoringNotificationChannel](/docs/catalog/gcp/gcpmonitoringnotificationchannel) — where incidents are delivered
- [GcpMonitoringUptimeCheck](/docs/catalog/gcp/gcpmonitoringuptimecheck) — the availability probe policies typically watch
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project and API enablement

## Additional Resources

- [Alerting Overview](https://cloud.google.com/monitoring/alerts)
- [AlertPolicy API Reference](https://cloud.google.com/monitoring/api/ref_v3/rest/v3/projects.alertPolicies)
- [Monitoring Filters](https://cloud.google.com/monitoring/api/v3/filters)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
