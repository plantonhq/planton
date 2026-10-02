# The Cloud Deploy API. disable_on_destroy is false: tearing down one
# policy must never disable the API for every pipeline in the project.
resource "google_project_service" "clouddeploy_api" {
  project = local.project_id
  service = "clouddeploy.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The deploy policy. Location and ID are immutable; the description,
# labels, annotations, suspension, rules, and selectors update in place.
# Clock and date parts are sent only when non-zero (an unset part is zero
# to Google), and empty lists and maps are not sent.
resource "google_clouddeploy_deploy_policy" "this" {
  project         = local.project_id
  location        = var.spec.location
  name            = local.deploy_policy_id
  description     = var.spec.description != "" ? var.spec.description : null
  labels          = local.final_labels
  annotations     = length(var.spec.annotations) > 0 ? var.spec.annotations : null
  suspended       = var.spec.suspended ? true : null
  deletion_policy = local.deletion_policy

  dynamic "rules" {
    for_each = var.spec.rules
    content {
      dynamic "rollout_restriction" {
        for_each = rules.value.rollout_restriction != null ? [rules.value.rollout_restriction] : []
        content {
          id       = rollout_restriction.value.id
          actions  = length(rollout_restriction.value.actions) > 0 ? rollout_restriction.value.actions : null
          invokers = length(rollout_restriction.value.invokers) > 0 ? rollout_restriction.value.invokers : null

          time_windows {
            time_zone = rollout_restriction.value.time_windows.time_zone

            dynamic "one_time_windows" {
              for_each = rollout_restriction.value.time_windows.one_time_windows
              content {
                start_date {
                  year  = one_time_windows.value.start_date.year != 0 ? one_time_windows.value.start_date.year : null
                  month = one_time_windows.value.start_date.month != 0 ? one_time_windows.value.start_date.month : null
                  day   = one_time_windows.value.start_date.day != 0 ? one_time_windows.value.start_date.day : null
                }
                start_time {
                  hours   = one_time_windows.value.start_time.hours != 0 ? one_time_windows.value.start_time.hours : null
                  minutes = one_time_windows.value.start_time.minutes != 0 ? one_time_windows.value.start_time.minutes : null
                  seconds = one_time_windows.value.start_time.seconds != 0 ? one_time_windows.value.start_time.seconds : null
                  nanos   = one_time_windows.value.start_time.nanos != 0 ? one_time_windows.value.start_time.nanos : null
                }
                end_date {
                  year  = one_time_windows.value.end_date.year != 0 ? one_time_windows.value.end_date.year : null
                  month = one_time_windows.value.end_date.month != 0 ? one_time_windows.value.end_date.month : null
                  day   = one_time_windows.value.end_date.day != 0 ? one_time_windows.value.end_date.day : null
                }
                end_time {
                  hours   = one_time_windows.value.end_time.hours != 0 ? one_time_windows.value.end_time.hours : null
                  minutes = one_time_windows.value.end_time.minutes != 0 ? one_time_windows.value.end_time.minutes : null
                  seconds = one_time_windows.value.end_time.seconds != 0 ? one_time_windows.value.end_time.seconds : null
                  nanos   = one_time_windows.value.end_time.nanos != 0 ? one_time_windows.value.end_time.nanos : null
                }
              }
            }

            dynamic "weekly_windows" {
              for_each = rollout_restriction.value.time_windows.weekly_windows
              content {
                days_of_week = length(weekly_windows.value.days_of_week) > 0 ? weekly_windows.value.days_of_week : null

                dynamic "start_time" {
                  for_each = weekly_windows.value.start_time != null ? [weekly_windows.value.start_time] : []
                  content {
                    hours   = start_time.value.hours != 0 ? start_time.value.hours : null
                    minutes = start_time.value.minutes != 0 ? start_time.value.minutes : null
                    seconds = start_time.value.seconds != 0 ? start_time.value.seconds : null
                    nanos   = start_time.value.nanos != 0 ? start_time.value.nanos : null
                  }
                }

                dynamic "end_time" {
                  for_each = weekly_windows.value.end_time != null ? [weekly_windows.value.end_time] : []
                  content {
                    hours   = end_time.value.hours != 0 ? end_time.value.hours : null
                    minutes = end_time.value.minutes != 0 ? end_time.value.minutes : null
                    seconds = end_time.value.seconds != 0 ? end_time.value.seconds : null
                    nanos   = end_time.value.nanos != 0 ? end_time.value.nanos : null
                  }
                }
              }
            }
          }
        }
      }
    }
  }

  # Selector labels are match criteria, so they carry exactly what the
  # spec declares -- never the attribution labels.
  dynamic "selectors" {
    for_each = var.spec.selectors
    content {
      dynamic "delivery_pipeline" {
        for_each = selectors.value.delivery_pipeline != null ? [selectors.value.delivery_pipeline] : []
        content {
          id     = delivery_pipeline.value.id != "" ? delivery_pipeline.value.id : null
          labels = length(delivery_pipeline.value.labels) > 0 ? delivery_pipeline.value.labels : null
        }
      }

      dynamic "target" {
        for_each = selectors.value.target != null ? [selectors.value.target] : []
        content {
          id     = target.value.id != "" ? target.value.id : null
          labels = length(target.value.labels) > 0 ? target.value.labels : null
        }
      }
    }
  }

  depends_on = [google_project_service.clouddeploy_api]
}
