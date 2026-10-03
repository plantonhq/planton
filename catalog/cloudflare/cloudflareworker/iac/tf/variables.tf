variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "CloudflareWorker specification"
  type = object({
    account_id = string
    worker_name = string
    compatibility_date = optional(string, "")
    content = optional(string, "")
    r2_bundle = optional(object({
      bucket = string
      path = string
    }))
    main_module = optional(string, "")
    compatibility_flags = optional(list(string), [])
    vars = optional(map(string), {})
    secrets = optional(list(object({
      name = string
      value = string
    })), [])
    kv_namespaces = optional(list(object({
      name = string
      namespace_id = string
    })), [])
    r2_buckets = optional(list(object({
      name = string
      bucket_name = string
      jurisdiction = optional(string, "")
    })), [])
    d1_databases = optional(list(object({
      name = string
      database_id = string
    })), [])
    hyperdrive_configs = optional(list(object({
      name = string
      config_id = string
    })), [])
    services = optional(list(object({
      name = string
      service = string
      environment = optional(string, "")
      entrypoint = optional(string, "")
    })), [])
    queues = optional(list(object({
      name = string
      queue_name = string
    })), [])
    durable_objects = optional(list(object({
      name = string
      class_name = string
      script_name = optional(string, "")
      environment = optional(string, "")
      namespace_id = optional(string, "")
      dispatch_namespace = optional(string, "")
    })), [])
    analytics_engine_datasets = optional(list(object({
      name = string
      dataset = string
    })), [])
    vectorize_indexes = optional(list(object({
      name = string
      index_name = string
    })), [])
    ai = optional(list(object({
      name = string
    })), [])
    version_metadata = optional(list(object({
      name = string
    })), [])
    workers_dev = optional(object({
      enabled = optional(bool, false)
      previews_enabled = optional(bool, false)
    }))
    custom_domains = optional(list(object({
      hostname = string
      zone_id = optional(string, "")
    })), [])
    routes = optional(list(object({
      zone_id = string
      pattern = string
    })), [])
    schedules = optional(list(string), [])
    observability = optional(object({
      enabled = optional(bool, false)
      head_sampling_rate = optional(number, 0)
      logs = optional(object({
        enabled = optional(bool, false)
        invocation_logs = optional(bool, false)
        destinations = optional(list(string), [])
        head_sampling_rate = optional(number, 0)
        persist = optional(bool, false)
      }))
      traces = optional(object({
        destinations = optional(list(string), [])
        enabled = optional(bool, false)
        head_sampling_rate = optional(number, 0)
        persist = optional(bool, false)
        propagation_policy = optional(string, "")
      }))
    }))
    placement = optional(object({
      mode = optional(string, "")
    }))
    limits = optional(object({
      cpu_ms = optional(number, 0)
      subrequests = optional(number, 0)
    }))
    logpush = optional(bool, false)
    tail_consumers = optional(list(object({
      service = string
      environment = optional(string, "")
      namespace = optional(string, "")
    })), [])
    assets = optional(object({
      directory = string
      config = optional(object({
        html_handling = optional(string, "")
        not_found_handling = optional(string, "")
        headers = optional(string, "")
        redirects = optional(string, "")
        run_worker_first = optional(bool, false)
        run_worker_first_rules = optional(list(string), [])
      }))
      binding_name = optional(string, "")
    }))
    migrations = optional(object({
      deleted_classes = optional(list(string), [])
      new_classes = optional(list(string), [])
      new_sqlite_classes = optional(list(string), [])
      new_tag = optional(string, "")
      old_tag = optional(string, "")
      renamed_classes = optional(list(object({
        from = optional(string, "")
        to = optional(string, "")
      })), [])
      transferred_classes = optional(list(object({
        from = optional(string, "")
        from_script = optional(string, "")
        to = optional(string, "")
      })), [])
      steps = optional(list(object({
        deleted_classes = optional(list(string), [])
        new_classes = optional(list(string), [])
        new_sqlite_classes = optional(list(string), [])
        renamed_classes = optional(list(object({
          from = optional(string, "")
          to = optional(string, "")
        })), [])
        transferred_classes = optional(list(object({
          from = optional(string, "")
          from_script = optional(string, "")
          to = optional(string, "")
        })), [])
      })), [])
    }))
    keep_assets = optional(bool, false)
    keep_bindings = optional(list(string), [])
    usage_model = optional(string, "")
    cache_options = optional(object({
      enabled = optional(bool, false)
      cross_version_cache = optional(bool, false)
    }))
    exports = optional(map(object({
      type = string
      cache = optional(object({
        enabled = optional(bool, false)
      }))
    })), {})
    package_dependencies = optional(list(object({
      name = string
      installed_version = string
      package_json_version = string
    })), [])
    annotations = optional(object({
      workers_message = optional(string, "")
      workers_tag = optional(string, "")
    }))
    body_part = optional(string, "")
    content_type = optional(string, "")
    mtls_certificates = optional(list(object({
      name = string
      certificate_id = string
    })), [])
    dispatch_namespaces = optional(list(object({
      name = string
      namespace = string
      outbound = optional(object({
        params = optional(list(string), [])
        worker = optional(object({
          service = optional(string, "")
          environment = optional(string, "")
        }))
      }))
    })), [])
    rate_limits = optional(list(object({
      name = string
      namespace = string
      simple = object({
        limit = number
        period = number
        mitigation_timeout = optional(number, 0)
      })
    })), [])
    send_email = optional(list(object({
      name = string
      destination_address = optional(string, "")
      allowed_destination_addresses = optional(list(string), [])
      allowed_sender_addresses = optional(list(string), [])
    })), [])
    secrets_store_secrets = optional(list(object({
      name = string
      store_id = string
      secret_name = string
    })), [])
    secret_keys = optional(list(object({
      name = string
      algorithm = string
      format = string
      usages = optional(list(string), [])
      key_base64 = optional(string, "")
      key_jwk = optional(string, "")
    })), [])
    workflows = optional(list(object({
      name = string
      workflow_name = string
    })), [])
    pipelines = optional(list(object({
      name = string
      pipeline = string
    })), [])
    json_bindings = optional(list(object({
      name = string
      json = string
    })), [])
    inherit_bindings = optional(list(object({
      name = string
      old_name = optional(string, "")
      version_id = optional(string, "")
    })), [])
    data_blobs = optional(list(object({
      name = string
      part = string
    })), [])
    text_blobs = optional(list(object({
      name = string
      part = string
    })), [])
    browsers = optional(list(object({
      name = string
    })), [])
    ai_search = optional(list(object({
      name = string
      instance_name = string
      namespace = optional(string, "")
      app_id = optional(string, "")
    })), [])
    ai_search_namespaces = optional(list(object({
      name = string
      namespace = string
    })), [])
    images = optional(list(object({
      name = string
    })), [])
    media = optional(list(object({
      name = string
    })), [])
    wasm_modules = optional(list(object({
      name = string
      part = string
    })), [])
    vpc_services = optional(list(object({
      name = string
      service_id = string
    })), [])
    vpc_networks = optional(list(object({
      name = string
      network_id = optional(string, "")
      tunnel_id = optional(string, "")
    })), [])
    tail_consumer_bindings = optional(list(object({
      name = string
      service = string
    })), [])
  })
}