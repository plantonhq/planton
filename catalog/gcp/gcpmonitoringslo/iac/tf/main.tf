# Enable the Cloud Monitoring API so a fresh project can host the SLO.
# disable_on_destroy is false: tearing down one SLO must never disable
# monitoring for everything else in the project.
resource "google_project_service" "monitoring_api" {
  project = local.project_id
  service = "monitoring.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The custom Monitoring service the SLO measures, when the spec's service
# arm asks for one (one kind, up to two service resources, count-gated —
# exactly one of the three arms is set, spec-validated). The service
# follows the kind's deletion contract: destroying the SLO component destroys
# the service it created.
resource "google_monitoring_custom_service" "this" {
  count = local.create_custom_service ? 1 : 0

  service_id   = local.created_service_id
  display_name = local.custom_service_display_name
  project      = local.project_id
  user_labels  = local.final_labels

  dynamic "telemetry" {
    for_each = var.spec.service.custom_service.telemetry_resource_name != "" ? [1] : []
    content {
      resource_name = var.spec.service.custom_service.telemetry_resource_name
    }
  }

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  depends_on = [google_project_service.monitoring_api]
}

# The basic (generic) Monitoring service — created from a well-known
# service type + identifying labels; GCP wires the telemetry association
# from the labels.
resource "google_monitoring_service" "this" {
  count = local.create_basic_service ? 1 : 0

  service_id   = local.created_service_id
  display_name = local.display_name
  project      = local.project_id
  user_labels  = local.final_labels

  basic_service {
    service_type   = var.spec.service.basic_service.service_type != "" ? var.spec.service.basic_service.service_type : null
    service_labels = length(var.spec.service.basic_service.service_labels) > 0 ? var.spec.service.basic_service.service_labels : null
  }

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  depends_on = [google_project_service.monitoring_api]
}

# The service-level objective. Exactly one SLI family and exactly one
# period form are set (spec-validated, mirroring the provider's
# ExactlyOneOf), so each dynamic block below renders at most once.
#
# availability.enabled is sent EXPLICITLY where the arm is present: it is
# Optional in the provider, and the GCP API expects true — omitting it
# would leave the arm's one field to a server-side default the spec cannot
# see.
resource "google_monitoring_slo" "this" {
  service      = local.slo_service_id
  goal         = var.spec.goal
  display_name = local.display_name
  project      = local.project_id
  user_labels  = local.final_labels

  slo_id = var.spec.slo_id != "" ? var.spec.slo_id : null

  calendar_period     = var.spec.calendar_period != "" ? var.spec.calendar_period : null
  rolling_period_days = var.spec.rolling_period_days != 0 ? var.spec.rolling_period_days : null

  dynamic "basic_sli" {
    for_each = var.spec.sli.basic_sli != null ? [var.spec.sli.basic_sli] : []
    content {
      location = length(basic_sli.value.location) > 0 ? basic_sli.value.location : null
      method   = length(basic_sli.value.method) > 0 ? basic_sli.value.method : null
      version  = length(basic_sli.value.version) > 0 ? basic_sli.value.version : null

      dynamic "availability" {
        for_each = basic_sli.value.availability != null ? [basic_sli.value.availability] : []
        content {
          # Explicit send — see the resource comment.
          enabled = availability.value.enabled == null ? true : availability.value.enabled
        }
      }

      dynamic "latency" {
        for_each = basic_sli.value.latency != null ? [basic_sli.value.latency] : []
        content {
          threshold = latency.value.threshold
        }
      }
    }
  }

  dynamic "request_based_sli" {
    for_each = var.spec.sli.request_based_sli != null ? [var.spec.sli.request_based_sli] : []
    content {
      dynamic "distribution_cut" {
        for_each = request_based_sli.value.distribution_cut != null ? [request_based_sli.value.distribution_cut] : []
        content {
          distribution_filter = distribution_cut.value.distribution_filter

          dynamic "range" {
            for_each = distribution_cut.value.range != null ? [distribution_cut.value.range] : []
            content {
              # Unset bounds stay null — the API treats them as unbounded
              # (0 is a meaningful bound, distinct from "no bound").
              min = range.value.min
              max = range.value.max
            }
          }
        }
      }

      dynamic "good_total_ratio" {
        for_each = request_based_sli.value.good_total_ratio != null ? [request_based_sli.value.good_total_ratio] : []
        content {
          good_service_filter  = good_total_ratio.value.good_service_filter != "" ? good_total_ratio.value.good_service_filter : null
          bad_service_filter   = good_total_ratio.value.bad_service_filter != "" ? good_total_ratio.value.bad_service_filter : null
          total_service_filter = good_total_ratio.value.total_service_filter != "" ? good_total_ratio.value.total_service_filter : null
        }
      }
    }
  }

  dynamic "windows_based_sli" {
    for_each = var.spec.sli.windows_based_sli != null ? [var.spec.sli.windows_based_sli] : []
    content {
      window_period          = windows_based_sli.value.window_period != "" ? windows_based_sli.value.window_period : null
      good_bad_metric_filter = windows_based_sli.value.good_bad_metric_filter != "" ? windows_based_sli.value.good_bad_metric_filter : null

      dynamic "good_total_ratio_threshold" {
        for_each = windows_based_sli.value.good_total_ratio_threshold != null ? [windows_based_sli.value.good_total_ratio_threshold] : []
        content {
          # threshold 0 is a legal (if degenerate) ratio bound, so it is
          # always sent — zero and unset are deliberately NOT distinguished
          # for this field.
          threshold = good_total_ratio_threshold.value.threshold

          dynamic "basic_sli_performance" {
            for_each = good_total_ratio_threshold.value.basic_sli_performance != null ? [good_total_ratio_threshold.value.basic_sli_performance] : []
            content {
              location = length(basic_sli_performance.value.location) > 0 ? basic_sli_performance.value.location : null
              method   = length(basic_sli_performance.value.method) > 0 ? basic_sli_performance.value.method : null
              version  = length(basic_sli_performance.value.version) > 0 ? basic_sli_performance.value.version : null

              dynamic "availability" {
                for_each = basic_sli_performance.value.availability != null ? [basic_sli_performance.value.availability] : []
                content {
                  enabled = availability.value.enabled == null ? true : availability.value.enabled
                }
              }

              dynamic "latency" {
                for_each = basic_sli_performance.value.latency != null ? [basic_sli_performance.value.latency] : []
                content {
                  threshold = latency.value.threshold
                }
              }
            }
          }

          dynamic "performance" {
            for_each = good_total_ratio_threshold.value.performance != null ? [good_total_ratio_threshold.value.performance] : []
            content {
              dynamic "distribution_cut" {
                for_each = performance.value.distribution_cut != null ? [performance.value.distribution_cut] : []
                content {
                  distribution_filter = distribution_cut.value.distribution_filter

                  dynamic "range" {
                    for_each = distribution_cut.value.range != null ? [distribution_cut.value.range] : []
                    content {
                      min = range.value.min
                      max = range.value.max
                    }
                  }
                }
              }

              dynamic "good_total_ratio" {
                for_each = performance.value.good_total_ratio != null ? [performance.value.good_total_ratio] : []
                content {
                  good_service_filter  = good_total_ratio.value.good_service_filter != "" ? good_total_ratio.value.good_service_filter : null
                  bad_service_filter   = good_total_ratio.value.bad_service_filter != "" ? good_total_ratio.value.bad_service_filter : null
                  total_service_filter = good_total_ratio.value.total_service_filter != "" ? good_total_ratio.value.total_service_filter : null
                }
              }
            }
          }
        }
      }

      dynamic "metric_mean_in_range" {
        for_each = windows_based_sli.value.metric_mean_in_range != null ? [windows_based_sli.value.metric_mean_in_range] : []
        content {
          time_series = metric_mean_in_range.value.time_series

          # The provider REQUIRES a range block on mean/sum criteria; a nil
          # spec range still renders an empty block and lets the API report
          # the miss precisely.
          range {
            min = metric_mean_in_range.value.range != null ? metric_mean_in_range.value.range.min : null
            max = metric_mean_in_range.value.range != null ? metric_mean_in_range.value.range.max : null
          }
        }
      }

      dynamic "metric_sum_in_range" {
        for_each = windows_based_sli.value.metric_sum_in_range != null ? [windows_based_sli.value.metric_sum_in_range] : []
        content {
          time_series = metric_sum_in_range.value.time_series

          range {
            min = metric_sum_in_range.value.range != null ? metric_sum_in_range.value.range.min : null
            max = metric_sum_in_range.value.range != null ? metric_sum_in_range.value.range.max : null
          }
        }
      }
    }
  }

  # Empty defers to the provider default (DELETE).
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  depends_on = [google_project_service.monitoring_api]
}
