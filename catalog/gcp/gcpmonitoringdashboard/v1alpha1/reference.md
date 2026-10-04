# GcpMonitoringDashboard

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpMonitoringDashboardSpec defines a Cloud Monitoring dashboard — the
console page of charts, scorecards, tables, and text widgets a team
watches to understand a service's health at a glance.

The dashboard body is ONE JSON document (`dashboard_json`) in the
Monitoring API's Dashboard format. That is a deliberate design mirroring
the GCP Terraform provider's own judgment: the Dashboard object is an
enormous, fast-moving schema (dozens of widget types, each with its own
options), and the provider models it as a single JSON-string argument
rather than inventing a structure that would forever lag the API. This
kind does the same — presets carry complete, working dashboards to start
from, and any dashboard built in the GCP console can be exported as JSON
and pasted here verbatim.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpMonitoringDashboard
metadata:
  name: my-sample-dashboard
spec:
  # The whole dashboard as ONE JSON document in the Monitoring API's
  # Dashboard format (https://cloud.google.com/monitoring/api/ref_v3/rest/v1/projects.dashboards):
  # displayName plus exactly one layout (gridLayout / mosaicLayout /
  # rowLayout / columnLayout) whose widgets carry the charts. Build
  # visually in the GCP console, then JSON-export (dashboard settings ->
  # "JSON editor") and paste the document here — server-assigned keys
  # round-trip cleanly.
  dashboardJson: |
    {
      "displayName": "API health",
      "gridLayout": {
        "columns": "2",
        "widgets": [
          {
            "title": "CPU utilization",
            "xyChart": {
              "dataSets": [
                {
                  "timeSeriesQuery": {
                    "timeSeriesFilter": {
                      "filter": "metric.type=\"compute.googleapis.com/instance/cpu/utilization\" resource.type=\"gce_instance\"",
                      "aggregation": {
                        "alignmentPeriod": "60s",
                        "perSeriesAligner": "ALIGN_MEAN"
                      }
                    }
                  }
                }
              ]
            }
          },
          {
            "title": "Notes",
            "text": {
              "content": "Managed by Planton — edit the manifest, not the console.",
              "format": "MARKDOWN"
            }
          }
        ]
      }
    }

  # What a destroy does: DELETE (default; dashboards are views — the
  # metrics keep flowing), PREVENT, or ABANDON.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.dashboardJson` | `string` | yes |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the dashboard. Can be a literal project ID or
a reference to a GcpProject resource. If omitted, the provider's
default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.dashboardJson

`string` · required

The dashboard as one JSON document, following the Monitoring API's
Dashboard format:
https://cloud.google.com/monitoring/api/ref_v3/rest/v1/projects.dashboards

The top-level keys are `displayName` plus exactly one layout —
`gridLayout`, `mosaicLayout`, `rowLayout`, or `columnLayout` — whose
widgets carry the charts. A minimal one-chart dashboard:

  {
    "displayName": "API health",
    "gridLayout": {
      "columns": "2",
      "widgets": [{
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
      }]
    }
  }

The document must be valid JSON (checked at plan time). Server-assigned
keys (etag, name) are ignored on the way back in, so a dashboard
exported from the console round-trips cleanly. To build visually: edit
the dashboard in the GCP console, then JSON-export it (dashboard
settings -> "JSON editor") and paste the document here.

- rule: {"required":true}

### spec.deletionPolicy

`string`

Deletion policy — what happens when this resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the dashboard is deleted from the project
  "PREVENT" -- destroy FAILS; protects a team's primary operational
               view from accidental teardown
  "ABANDON" -- the dashboard is removed from management but stays
               visible in the GCP console

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpMonitoringDashboard, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.dashboard_name` | `string` | The server-assigned resource name of the dashboard. Format: projects/{project}/dashboards/{dashboard_id} The handle for linking to the dashboard from alert documentation and for addressing it through the Monitoring dashboards API. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## See Also

- [Overview](../README.md)
