# GCP Monitoring Uptime Check

Creates a Cloud Monitoring uptime check — a probe Google runs against your target (a public URL, a monitored resource, a resource group, or a synthetic-monitor Cloud Function) from multiple regions on a fixed cadence, recording availability and latency as metrics. Pair it with a GcpMonitoringAlertPolicy on the check's failure metric and the "is my site up" question becomes one declarative graph.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Uptime Check Config** -- a `monitoring.UptimeCheckConfig` with the configured target, probe (HTTP/HTTPS or TCP), cadence, regions, and content assertions
- **Monitoring API enablement** -- `monitoring.googleapis.com` enabled in the target project (never disabled on destroy)
- **GCP Labels** -- resource metadata labels applied automatically as `user_labels` for tracking and governance

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery.

### GCP Project

- **A GCP project** where the check is created (directly or via a GcpProject reference).
- **For synthetic monitors**: a 2nd-gen Cloud Function carrying the probe logic (reference a GcpCloudFunction resource).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpMonitoringUptimeCheck
metadata:
  name: website-https
spec:
  timeout: 10s
  monitoredResource:
    type: uptime_url
    labels:
      host: example.com
  httpCheck:
    path: /
    useSsl: true
    validateSsl: true
```

```shell
planton apply -f uptime-check.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `timeout` | `string` | Max wait for the probe (e.g. `10s`). | 1s–60s, whole seconds |
| one target | `message` | Exactly one of `monitoredResource`, `resourceGroup`, `syntheticMonitor`. | Exactly-one |
| one check | `message` | Exactly one of `httpCheck`, `tcpCheck` (omit both only for a synthetic monitor). | Exactly-one |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default | GCP project. Can reference a GcpProject resource. |
| `displayName` | `string` | `metadata.name` | Console display name. |
| `period` | `string` | `300s` | Check cadence: `60s`, `300s`, `600s`, or `900s` — the only values GCP accepts. |
| `checkerType` | `string` | `STATIC_IP_CHECKERS` | `STATIC_IP_CHECKERS` (public fleet) or `VPC_CHECKERS` (private targets). |
| `selectedRegions` | `list<string>` | all regions | Probe origins; must cover ≥3 checker locations when set. |
| `logCheckFailures` | `bool` | `false` | Log failed probes to Cloud Logging. |
| `contentMatchers` | `list` | `[]` | Response-body assertions (string/regex/JSON-path). |
| `labels` | `map<string,string>` | `{}` | User metadata labels, merged with platform labels. |
| `deletionPolicy` | `string` | `DELETE` | What destroy does: `DELETE`, `PREVENT`, `ABANDON`. |

Key nested surfaces: `httpCheck` (path, port, request method, TLS + certificate validation, base64 body + content type, headers + masking, basic auth — the password is a managed secret — service-agent OIDC auth, accepted status codes, pings) and `tcpCheck` (port, pings).

### Validation Rules

- **Exactly one target; exactly one check** (none with a synthetic monitor — the function carries its own probe logic).
- **`customContentType` pairs with `contentType: USER_PROVIDED`** in both directions.
- **JSON-path sub-matcher** is required with (and only with) the `MATCHES_JSON_PATH` / `NOT_MATCHES_JSON_PATH` matchers.
- **Status-code entries** set a class or an exact value, never both.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `uptime_check_name` | `string` | `projects/{project}/uptimeCheckConfigs/{id}` |
| `uptime_check_id` | `string` | The bare check ID — the value an alert policy's filter references as `metric.label.check_id` |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md).

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md).

## Important Notes

- **A check only measures.** To be paged, pair it with a GcpMonitoringAlertPolicy whose threshold condition filters on `monitoring.googleapis.com/uptime_check/check_passed` and this check's `uptime_check_id`.
- **`validateSsl` defaults to false** — enable it in production so an expired certificate fails the probe instead of passing silently.
- **`maskHeaders` is permanent once enabled** — turning it back off requires recreating the check.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- [GcpMonitoringAlertPolicy](/docs/catalog/gcp/gcpmonitoringalertpolicy) — pages when this check fails
- [GcpMonitoringNotificationChannel](/docs/catalog/gcp/gcpmonitoringnotificationchannel) — where those pages are delivered
- [GcpCloudFunction](/docs/catalog/gcp/gcpcloudfunction) — carries the probe logic for synthetic monitors

## Additional Resources

- [Uptime Checks Overview](https://cloud.google.com/monitoring/uptime-checks)
- [UptimeCheckConfig API Reference](https://cloud.google.com/monitoring/api/ref_v3/rest/v3/projects.uptimeCheckConfigs)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
