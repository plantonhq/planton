# GCP Monitoring Dashboard

Creates a Cloud Monitoring dashboard — the console page of charts, scorecards, tables, and text widgets a team watches to understand a service's health at a glance. The dashboard body is one JSON document in the Monitoring API's own format: build visually in the GCP console, export the JSON, and manage it declaratively from then on.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Dashboard** -- a `monitoring.Dashboard` from the spec's `dashboardJson` document
- **Monitoring API enablement** -- `monitoring.googleapis.com` enabled in the target project (never disabled on destroy)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery.

### GCP Project

- **A GCP project** where the dashboard is created (directly or via a GcpProject reference).
- **IAM**: the deploying identity needs `roles/monitoring.dashboardEditor` or broader.

## Deploy

### Console

Open the deployment store, find **GCP Monitoring Dashboard**, and click **Deploy**. The creation wizard walks you through the target project and the dashboard JSON document. Start from the **Golden Signals** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpMonitoringDashboard
metadata:
  name: api-health
  org: acme-corp
  env: prod
spec:
  dashboardJson: |
    {
      "displayName": "API health",
      "gridLayout": {
        "columns": "2",
        "widgets": [
          {
            "title": "CPU utilization",
            "xyChart": {
              "dataSets": [{
                "timeSeriesQuery": {
                  "timeSeriesFilter": {
                    "filter": "metric.type=\"compute.googleapis.com/instance/cpu/utilization\" resource.type=\"gce_instance\"",
                    "aggregation": {"perSeriesAligner": "ALIGN_MEAN"}
                  }
                }
              }]
            }
          }
        ]
      }
    }
```

```shell
planton apply -f dashboard.yaml
```

This creates a one-chart grid dashboard charting fleet CPU utilization. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the dashboard references its project via ValueFromRef:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: observability-project
      fieldPath: status.outputs.project_id
  dashboardJson: |
    { "displayName": "API health", "gridLayout": { "columns": "2", "widgets": [] } }
```

The InfraPipeline deploys the project first, then creates the dashboard in it.

## Key Configuration

These are the most important decisions when configuring a dashboard. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**dashboardJson** -- the whole dashboard as one JSON document in the Monitoring API's own Dashboard format: `displayName` plus exactly one layout (`gridLayout`, `mosaicLayout`, `rowLayout`, or `columnLayout`) whose widgets carry the charts. The practical workflow: build the dashboard visually in the GCP console, open its **JSON editor**, and paste the export here — server-assigned keys round-trip cleanly.

**deletionPolicy** -- `DELETE` (default), `PREVENT` (protects a team's primary operational view from accidental teardown), or `ABANDON` (remove from management, keep in GCP).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `dashboard_name` | `projects/{project}/dashboards/{id}` | Alert-documentation links, Monitoring API cross-references |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Golden signals** -- the SRE starter dashboard: traffic, errors, latency, saturation on one page. Start from the **Golden Signals** preset.

**Infrastructure overview** -- fleet-level CPU, memory, disk, and network for Compute Engine workloads. Start from the **Infrastructure Overview** preset.

## Works With

- [**GCP Monitoring Alert Policy**](/infra-catalog/gcp-monitoring-alert-policy) -- link the dashboard from alert runbook documentation
- [**GCP Log Metric**](/infra-catalog/gcp-log-metric) -- chart log-based metrics on dashboard widgets
- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the dashboard is created
