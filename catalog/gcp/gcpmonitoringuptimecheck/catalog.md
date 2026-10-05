# GCP Monitoring Uptime Check

Creates a Cloud Monitoring uptime check — a probe Google runs against your target (a public URL, a monitored resource, a resource group, or a synthetic-monitor Cloud Function) from multiple regions on a fixed cadence, recording availability and latency as metrics. Pair it with a GCP Monitoring Alert Policy on the check's failure metric and the "is my site up" question becomes one declarative graph.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Uptime Check Config** -- a `monitoring.UptimeCheckConfig` with the configured target, probe (HTTP/HTTPS or TCP), cadence, regions, and content assertions
- **Monitoring API enablement** -- `monitoring.googleapis.com` enabled in the target project (never disabled on destroy)
- **GCP Labels** -- resource metadata labels applied automatically as `user_labels`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery.

### GCP Project

- **A GCP project** where the check is created (directly or via a GcpProject reference).
- **IAM**: the deploying identity needs `roles/monitoring.editor` or broader.

## Deploy

### Console

Open the deployment store, find **GCP Monitoring Uptime Check**, and click **Deploy**. The creation wizard walks you through the target project, the probe target, the HTTP/TCP check settings, cadence and regions, and content assertions. Start from the **Public HTTPS Check** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpMonitoringUptimeCheck
metadata:
  name: website-https
  org: acme-corp
  env: prod
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

This probes `https://example.com/` from all regions every 5 minutes (the default period), failing on non-2xx responses and invalid certificates. An Infra Job tracks the provisioning in real time.

### InfraChart

For a synthetic monitor, the check references its probe function via ValueFromRef:

```yaml
spec:
  syntheticMonitor:
    cloudFunction:
      valueFrom:
        kind: GcpCloudFunction
        name: checkout-synthetic
        fieldPath: status.outputs.function_id
```

The InfraPipeline deploys the function first, then the check against it — and an alert policy in the same chart pages on the check's `check_passed` metric, closing the paging path.

## Key Configuration

These are the most important decisions when configuring an uptime check. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Target** -- exactly one of `monitoredResource` (the canonical public-URL form: type `uptime_url` with a `host` label), `resourceGroup` (probe every member), or `syntheticMonitor` (a 2nd-gen Cloud Function carries the probe logic).

**Check** -- exactly one of `httpCheck` (path, TLS, auth, expected status codes, body assertions) or `tcpCheck` (port connect). Synthetic monitors take neither.

**Cadence and coverage** -- `period` accepts only 60s/300s/600s/900s; leaving `selectedRegions` empty probes from ALL regions — the recommended default.

**Content matchers** -- assert on the response body (contains/regex/JSON-path) so a 200 serving an error page still fails the probe.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpCloudFunction** (optional) | `syntheticMonitor.cloudFunction` | `status.outputs.function_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `uptime_check_id` | The bare check ID | GcpMonitoringAlertPolicy threshold filters (`metric.label.check_id`) — the check-plus-alert composition key |
| `uptime_check_name` | Full resource name | Monitoring API cross-references |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Public HTTPS check** -- probe a public URL over TLS with certificate validation — the canonical availability monitor. Start from the **Public HTTPS Check** preset.

**TCP port check** -- assert a non-HTTP service (a database endpoint, a message broker) accepts connections. Start from the **TCP Port Check** preset.

**Authenticated API check** -- probe an endpoint that requires basic auth, accepting 401-free responses only, with a JSON body assertion. Start from the **Authenticated API Check** preset.

## Works With

- [**GCP Monitoring Alert Policy**](/infra-catalog/gcp-monitoring-alert-policy) -- pages when this check fails
- [**GCP Monitoring Notification Channel**](/infra-catalog/gcp-monitoring-notification-channel) -- where those pages are delivered
- [**GCP Cloud Function**](/infra-catalog/gcp-cloud-function) -- carries the probe logic for synthetic monitors
