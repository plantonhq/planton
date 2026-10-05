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
- **IAM**: see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying identity needs.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpMonitoringDashboard
metadata:
  name: api-health
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

## Why one JSON field?

The Dashboard API object is an enormous, fast-moving schema — dozens of widget types, each with its own options. The GCP Terraform provider models it as a single JSON-string argument rather than inventing a structure that would forever lag the API, and this kind honors that judgment. The workflow that makes it pleasant: edit the dashboard in the GCP console, open its **JSON editor**, and paste the exported document into `dashboardJson`. Server-assigned keys (etag, name) are ignored on the way back in, so exported dashboards round-trip cleanly with no drift.

## Outputs

| Output | Description |
|--------|-------------|
| `dashboard_name` | `projects/{project}/dashboards/{id}` — link target for alert documentation, Monitoring API handle |

## Works With

- **GcpMonitoringAlertPolicy** -- link the dashboard from alert runbook documentation so the on-call engineer lands on the right charts
- **GcpLogMetric** -- chart log-based metrics (`logging.googleapis.com/user/{metric_name}`) on dashboard widgets
- **GcpProject** -- provides the GCP project where the dashboard is created

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
