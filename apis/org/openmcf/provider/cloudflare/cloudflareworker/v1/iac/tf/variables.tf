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
  description = "Specification for a Cloudflare Worker (script, bindings, routing, schedules, settings)"
  # NOTE: StringValueOrRef fields (binding ids, secret values, zone ids) are
  # flattened to plain strings by the proto->tfvars converter.
  type = object({
    account_id  = string
    worker_name = string

    compatibility_date  = optional(string, "")
    content             = optional(string, "")
    r2_bundle           = optional(object({ bucket = string, path = string }))
    main_module         = optional(string, "index.js")
    compatibility_flags = optional(list(string), [])

    # Bindings, grouped by type (flattened into the provider's bindings list).
    vars    = optional(map(string), {})
    secrets = optional(list(object({ name = string, value = string })), [])
    kv_namespaces = optional(list(object({
      name         = string
      namespace_id = optional(string, "")
    })), [])
    r2_buckets = optional(list(object({
      name         = string
      bucket_name  = optional(string, "")
      jurisdiction = optional(string, "")
    })), [])
    d1_databases = optional(list(object({
      name        = string
      database_id = optional(string, "")
    })), [])
    hyperdrive_configs = optional(list(object({
      name      = string
      config_id = optional(string, "")
    })), [])
    services = optional(list(object({
      name        = string
      service     = optional(string, "")
      environment = optional(string, "")
    })), [])
    queues = optional(list(object({
      name       = string
      queue_name = string
    })), [])
    durable_objects = optional(list(object({
      name        = string
      class_name  = string
      script_name = optional(string, "")
      environment = optional(string, "")
    })), [])
    analytics_engine_datasets = optional(list(object({
      name    = string
      dataset = string
    })), [])
    vectorize_indexes = optional(list(object({
      name       = string
      index_name = string
    })), [])
    ai               = optional(list(object({ name = string })), [])
    version_metadata = optional(list(object({ name = string })), [])

    # Routing.
    workers_dev = optional(object({
      enabled          = optional(bool, false)
      previews_enabled = optional(bool, false)
    }))
    custom_domains = optional(list(object({
      hostname = string
    })), [])
    routes = optional(list(object({
      zone_id = optional(string, "")
      pattern = string
    })), [])

    # Cron schedules.
    schedules = optional(list(string), [])

    # Runtime settings.
    observability = optional(object({
      enabled            = optional(bool, false)
      head_sampling_rate = optional(number, 0)
    }))
    placement = optional(object({
      mode = optional(string, "")
    }))
    limits = optional(object({
      cpu_ms = optional(number, 0)
    }))
    logpush = optional(bool, false)
    tail_consumers = optional(list(object({
      service     = string
      environment = optional(string, "")
      namespace   = optional(string, "")
    })), [])
  })
}
