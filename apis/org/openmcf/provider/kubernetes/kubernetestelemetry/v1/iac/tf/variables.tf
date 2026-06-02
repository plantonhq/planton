variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Specification for KubernetesTelemetry"
  type = object({
    # Namespace the Telemetry resource is created in (resolved foreign key).
    namespace = string

    # Workload selector. When omitted (and target_refs is also omitted), the
    # configuration applies to all workloads in the namespace. Matched by label at
    # runtime by istiod; not an OpenMCF foreign key. Mutually exclusive with
    # target_refs.
    selector = optional(object({
      match_labels = optional(map(string))
    }))

    # Attaches the configuration to specific resources (Gateway, Service,
    # ServiceEntry) instead of selecting workloads by label. Mutually exclusive with
    # selector. Plain cross-resource references, not OpenMCF foreign keys.
    target_refs = optional(list(object({
      group     = optional(string)
      kind      = string
      name      = string
      namespace = optional(string)
    })))

    # Tracing configuration. Each entry can set the match, providers, sampling rate,
    # custom span tags, and span-reporting toggles.
    tracing = optional(list(object({
      match = optional(object({
        mode = optional(string)
      }))
      providers                  = optional(list(object({ name = string })))
      random_sampling_percentage = optional(number)
      disable_span_reporting     = optional(bool)
      # Custom span tags keyed by tag name; each value carries exactly one source.
      custom_tags = optional(map(object({
        literal = optional(object({
          value = string
        }))
        environment = optional(object({
          name          = string
          default_value = optional(string)
        }))
        header = optional(object({
          name          = string
          default_value = optional(string)
        }))
      })))
      enable_istio_tags                 = optional(bool)
      use_request_id_for_trace_sampling = optional(bool)
    })))

    # Metrics configuration. Each entry can choose providers, apply ordered
    # overrides, and set the TCP reporting interval.
    metrics = optional(list(object({
      providers = optional(list(object({ name = string })))
      overrides = optional(list(object({
        match = optional(object({
          metric        = optional(string)
          custom_metric = optional(string)
          mode          = optional(string)
        }))
        disabled = optional(bool)
        # Tag (dimension) operations keyed by tag name.
        tag_overrides = optional(map(object({
          operation = optional(string)
          value     = optional(string)
        })))
      })))
      reporting_interval = optional(string)
    })))

    # Access logging configuration. Each entry can choose providers, toggle logging,
    # and attach a CEL filter.
    access_logging = optional(list(object({
      match = optional(object({
        mode = optional(string)
      }))
      providers = optional(list(object({ name = string })))
      disabled  = optional(bool)
      filter = optional(object({
        expression = string
      }))
    })))
  })
}
