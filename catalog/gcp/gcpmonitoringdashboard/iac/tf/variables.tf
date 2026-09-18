variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "GcpMonitoringDashboard specification"
  type = object({
    # The GCP project that owns the dashboard. Can be a literal project ID or
    # a reference to a GcpProject resource. If omitted, the provider's
    # default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The dashboard as one JSON document, following the Monitoring API's
    # Dashboard format:
    # https://cloud.google.com/monitoring/api/ref_v3/rest/v1/projects.dashboards
    #
    # The top-level keys are `displayName` plus exactly one layout —
    # `gridLayout`, `mosaicLayout`, `rowLayout`, or `columnLayout` — whose
    # widgets carry the charts. A minimal one-chart dashboard:
    #
    #   {
    #     "displayName": "API health",
    #     "gridLayout": {
    #       "columns": "2",
    #       "widgets": [{
    #         "title": "CPU utilization",
    #         "xyChart": {
    #           "dataSets": [{
    #             "timeSeriesQuery": {
    #               "timeSeriesFilter": {
    #                 "filter": "metric.type=\"compute.googleapis.com/instance/cpu/utilization\" resource.type=\"gce_instance\"",
    #                 "aggregation": {"perSeriesAligner": "ALIGN_MEAN"}
    #               }
    #             }
    #           }]
    #         }
    #       }]
    #     }
    #   }
    #
    # The document must be valid JSON (checked at plan time). Server-assigned
    # keys (etag, name) are ignored on the way back in, so a dashboard
    # exported from the console round-trips cleanly. To build visually: edit
    # the dashboard in the GCP console, then JSON-export it (dashboard
    # settings -> "JSON editor") and paste the document here.
    dashboard_json = string

    # Deletion policy — what happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the dashboard is deleted from the project
    #   "PREVENT" -- destroy FAILS; protects a team's primary operational
    #                view from accidental teardown
    #   "ABANDON" -- the dashboard is removed from management but stays
    #                visible in the GCP console
    deletion_policy = optional(string, "")
  })
}
