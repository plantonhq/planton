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
  description = "GcpMonitoringSlo specification"
  type = object({
    # The GCP project that owns the SLO (and any service this kind creates).
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Monitoring service this SLO measures — exactly one arm. Changing
    # the resolved service REPLACES the SLO (the GCP API binds an SLO to its
    # service for life).
    service = object({
      # Measure an EXISTING Monitoring service — one GCP auto-detected (App
      # Engine, Istio canonical services, and friends) or one created outside
      # this kind. The service ID is the last segment of the service resource
      # name (projects/{project}/services/{service_id}).
      service_id = optional(string, "")

      # CREATE a custom service and measure it. The blank-slate form: a custom
      # service is just a named container for SLOs — the SLIs point at
      # whatever metrics define the service.
      custom_service = optional(object({
        # The service ID (the last segment of the service resource name).
        # Defaults to metadata.name. Changing it REPLACES the service and the
        # SLO under it.
        service_id = optional(string, "")

        # Display name shown in the Monitoring services list. Defaults to
        # metadata.name.
        display_name = optional(string, "")

        # The full resource name of the underlying workload this service
        # represents (e.g. //run.googleapis.com/projects/{p}/locations/{l}/services/{s})
        # — how the console links the service to its resource. Optional.
        telemetry_resource_name = optional(string, "")
      }))

      # CREATE a basic service from a service type + labels (e.g. a
      # CLOUD_RUN service named by its service_name and location) and measure
      # it. GCP wires the telemetry association from the labels.
      basic_service = optional(object({
        # The service ID (the last segment of the service resource name).
        # Defaults to metadata.name. Changing it REPLACES the service and the
        # SLO under it.
        service_id = optional(string, "")

        # The service type, e.g. APP_ENGINE, CLOUD_ENDPOINTS, CLUSTER_ISTIO,
        # ISTIO_CANONICAL_SERVICE, CLOUD_RUN. The GCP API validates the type
        # and its required labels server-side (the provider carries no
        # client-side list — new types appear with new products).
        service_type = optional(string, "")

        # The labels that identify the concrete service instance for the chosen
        # type — e.g. for CLOUD_RUN: {"service_name": "checkout",
        # "location": "us-central1"}. Which keys are required is defined by the
        # service type (server-validated). Changing labels REPLACES the service.
        service_labels = optional(map(string), {})
      }))
    })

    # The SLO identifier within the service (the last path segment of the
    # SLO resource name). Omit to let GCP assign one. Changing it REPLACES
    # the SLO.
    slo_id = optional(string, "")

    # Human-friendly name shown in the SLO list and on dashboards. Defaults
    # to metadata.name when left empty.
    display_name = optional(string, "")

    # The target fraction of good service the SLO demands, e.g. 0.999 for
    # "three nines". Must be greater than 0 and at most 0.9999 (the GCP
    # API's own bound — five nines and beyond are not accepted).
    goal = number

    # Measure over a calendar period: DAY, WEEK, FORTNIGHT, or MONTH. The
    # error budget resets at each period boundary. Set exactly one of this
    # and rolling_period_days.
    calendar_period = optional(string, "")

    # Measure over a rolling window of this many days (1 to 30 — the GCP
    # API's bounds). The classic SRE form is 28 or 30. Set exactly one of
    # this and calendar_period.
    rolling_period_days = optional(number, 0)

    # The service-level indicator — how good service is counted. Exactly one
    # SLI family.
    sli = object({
      # A basic SLI: GCP derives availability or latency from the service's
      # own telemetry — no filters to write. Only works for service types
      # whose telemetry GCP understands natively (App Engine, Cloud
      # Endpoints, Istio); custom services need request_based_sli instead.
      basic_sli = optional(object({
        # Narrow the SLI to these locations (e.g. specific regions). Empty means
        # all locations.
        location = optional(list(string), [])

        # Narrow the SLI to these RPC methods. Empty means all methods.
        method = optional(list(string), [])

        # Narrow the SLI to these API versions. Empty means all versions.
        version = optional(list(string), [])

        # Count availability (successful vs total requests) as good service.
        availability = optional(object({
          # Whether the availability SLI is enabled (the GCP API expects true —
          # the field exists for API-shape fidelity). Both IaC engines send the
          # value explicitly so behavior is identical regardless of engine.
          enabled = optional(bool)
        }))

        # Count requests faster than a threshold as good service.
        latency = optional(object({
          # The latency threshold below which a response counts as good, as a
          # duration string (e.g. "1s", "0.5s").
          threshold = string
        }))
      }))

      # A request-based SLI: good service counted from metric filters — a
      # good/total ratio or a latency-distribution cut. The workhorse form
      # for custom services.
      request_based_sli = optional(object({
        # Good service = values of a distribution metric falling inside a range
        # (the latency-SLO form: requests whose latency lands in [0, 500ms]).
        distribution_cut = optional(object({
          # A monitoring filter (https://cloud.google.com/monitoring/api/v3/filters)
          # selecting a DISTRIBUTION-valued metric, e.g. request latencies.
          # The filter MUST pin resource.type alongside metric.type: GCP validates
          # at create that every SLO filter parses to exactly ONE monitored
          # resource type and rejects anything broader (Error 400 "parses to N
          # resource types and must parse to 1").
          distribution_filter = string

          # The range of values counted as good (e.g. min 0, max 500 for "under
          # 500ms" when the metric is in milliseconds).
          range = optional(object({
            # The lower bound of the range. Unset means unbounded below.
            min = optional(number)

            # The upper bound of the range. Unset means unbounded above.
            max = optional(number)
          }))
        }))

        # Good service = ratio of a "good" counter to a "total" counter (or
        # good vs bad).
        good_total_ratio = optional(object({
          # A monitoring filter counting GOOD events (a DELTA metric of int64 or
          # double type).
          good_service_filter = optional(string, "")

          # A monitoring filter counting BAD events.
          bad_service_filter = optional(string, "")

          # A monitoring filter counting TOTAL events.
          total_service_filter = optional(string, "")
        }))
      }))

      # A windows-based SLI: time is divided into windows and each window is
      # judged good or bad as a whole (e.g. "a window is good when p95
      # latency stays under 500ms"). For "no bad minutes" style objectives.
      windows_based_sli = optional(object({
        # The window length, as a duration string between "60s" and "604800s"
        # (one minute to one week). Empty defers to the GCP API default (60s).
        window_period = optional(string, "")

        # A window is good when this filter's BOOLEAN metric is true throughout
        # the window. Like every SLO filter, it must pin resource.type
        # alongside metric.type (GCP requires exactly one monitored resource
        # type per filter).
        good_bad_metric_filter = optional(string, "")

        # A window is good when a request-based criterion (a good/total ratio
        # or a basic SLI) meets a threshold within the window.
        good_total_ratio_threshold = optional(object({
          # The window is good when the criterion's ratio meets or exceeds this
          # threshold (0 to 1).
          threshold = optional(number, 0)

          # Judge each window with a basic SLI (availability or latency derived
          # from service telemetry).
          basic_sli_performance = optional(object({
            # Narrow the SLI to these locations (e.g. specific regions). Empty means
            # all locations.
            location = optional(list(string), [])

            # Narrow the SLI to these RPC methods. Empty means all methods.
            method = optional(list(string), [])

            # Narrow the SLI to these API versions. Empty means all versions.
            version = optional(list(string), [])

            # Count availability (successful vs total requests) as good service.
            availability = optional(object({
              # Whether the availability SLI is enabled (the GCP API expects true —
              # the field exists for API-shape fidelity). Both IaC engines send the
              # value explicitly so behavior is identical regardless of engine.
              enabled = optional(bool)
            }))

            # Count requests faster than a threshold as good service.
            latency = optional(object({
              # The latency threshold below which a response counts as good, as a
              # duration string (e.g. "1s", "0.5s").
              threshold = string
            }))
          }))

          # Judge each window with a request-based criterion (distribution cut or
          # good/total ratio).
          performance = optional(object({
            # Distribution-cut criterion for the window.
            distribution_cut = optional(object({
              # A monitoring filter (https://cloud.google.com/monitoring/api/v3/filters)
              # selecting a DISTRIBUTION-valued metric, e.g. request latencies.
              # The filter MUST pin resource.type alongside metric.type: GCP validates
              # at create that every SLO filter parses to exactly ONE monitored
              # resource type and rejects anything broader (Error 400 "parses to N
              # resource types and must parse to 1").
              distribution_filter = string

              # The range of values counted as good (e.g. min 0, max 500 for "under
              # 500ms" when the metric is in milliseconds).
              range = optional(object({
                # The lower bound of the range. Unset means unbounded below.
                min = optional(number)

                # The upper bound of the range. Unset means unbounded above.
                max = optional(number)
              }))
            }))

            # Good/total ratio criterion for the window.
            good_total_ratio = optional(object({
              # A monitoring filter counting GOOD events (a DELTA metric of int64 or
              # double type).
              good_service_filter = optional(string, "")

              # A monitoring filter counting BAD events.
              bad_service_filter = optional(string, "")

              # A monitoring filter counting TOTAL events.
              total_service_filter = optional(string, "")
            }))
          }))
        }))

        # A window is good when the MEAN of a metric stays inside a range for
        # the window.
        metric_mean_in_range = optional(object({
          # A monitoring filter selecting the time series whose windowed mean (or
          # sum) is judged against `range`.
          time_series = string

          # The range the windowed value must stay inside for the window to count
          # as good.
          range = optional(object({
            # The lower bound of the range. Unset means unbounded below.
            min = optional(number)

            # The upper bound of the range. Unset means unbounded above.
            max = optional(number)
          }))
        }))

        # A window is good when the SUM of a metric stays inside a range for
        # the window.
        metric_sum_in_range = optional(object({
          # A monitoring filter selecting the time series whose windowed mean (or
          # sum) is judged against `range`.
          time_series = string

          # The range the windowed value must stay inside for the window to count
          # as good.
          range = optional(object({
            # The lower bound of the range. Unset means unbounded below.
            min = optional(number)

            # The upper bound of the range. Unset means unbounded above.
            max = optional(number)
          }))
        }))
      }))
    })

    # User labels attached to the SLO for organizing and identifying it
    # (maps to the provider's user_labels), merged with Planton's platform
    # labels (which win on key conflicts). Also applied to any Monitoring
    # service this kind creates. Keys and values may contain only lowercase
    # letters, numerals, underscores, and dashes; keys must begin with a
    # letter.
    labels = optional(map(string), {})

    # Deletion policy — what happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the SLO (and any service this kind created) is deleted;
    #                burn-rate alerts referencing it stop evaluating
    #   "PREVENT" -- destroy FAILS; protects the reliability contract from
    #                accidental teardown
    #   "ABANDON" -- the SLO is removed from management but keeps existing
    #                in GCP
    deletion_policy = optional(string, "")
  })
}
