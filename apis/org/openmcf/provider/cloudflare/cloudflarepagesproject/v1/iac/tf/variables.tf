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
  description = "Specification for a Cloudflare Pages project (build config, git source, per-environment deployment configs, custom domains)"
  # NOTE: StringValueOrRef fields (binding ids, secret values) are flattened to
  # plain strings by the proto->tfvars converter. The per-environment deployment
  # config shape is identical for preview and production.
  type = object({
    account_id        = string
    name              = string
    production_branch = string

    build_config = optional(object({
      build_command       = optional(string, "")
      destination_dir     = optional(string, "")
      root_dir            = optional(string, "")
      build_caching       = optional(bool, false)
      web_analytics_tag   = optional(string, "")
      web_analytics_token = optional(string, "")
    }))

    source = optional(object({
      type = string
      config = object({
        owner                          = optional(string, "")
        repo_name                      = optional(string, "")
        production_branch              = optional(string, "")
        pr_comments_enabled            = optional(bool, false)
        deployments_enabled            = optional(bool, false)
        production_deployments_enabled = optional(bool, false)
        preview_deployment_setting     = optional(string, "")
        preview_branch_includes        = optional(list(string), [])
        preview_branch_excludes        = optional(list(string), [])
        path_includes                  = optional(list(string), [])
        path_excludes                  = optional(list(string), [])
      })
    }))

    deployment_configs = optional(object({
      preview = optional(object({
        compatibility_date                   = optional(string, "")
        compatibility_flags                  = optional(list(string), [])
        always_use_latest_compatibility_date = optional(bool, false)
        build_image_major_version            = optional(number, 0)
        fail_open                            = optional(bool, false)
        usage_model                          = optional(string, "")
        limits                               = optional(object({ cpu_ms = optional(number, 0) }))
        placement                            = optional(object({ mode = optional(string, "") }))
        vars                                 = optional(map(string), {})
        secrets                              = optional(list(object({ name = string, value = string })), [])
        kv_namespaces                        = optional(list(object({ name = string, namespace_id = string })), [])
        d1_databases                         = optional(list(object({ name = string, database_id = string })), [])
        r2_buckets                           = optional(list(object({ name = string, bucket_name = string, jurisdiction = optional(string, "") })), [])
        queue_producers                      = optional(list(object({ name = string, queue_name = string })), [])
        hyperdrive_bindings                  = optional(list(object({ name = string, config_id = string })), [])
        services                             = optional(list(object({ name = string, service = string, entrypoint = optional(string, ""), environment = optional(string, "") })), [])
        durable_object_namespaces            = optional(list(object({ name = string, namespace_id = string })), [])
        analytics_engine_datasets            = optional(list(object({ name = string, dataset = string })), [])
        vectorize_bindings                   = optional(list(object({ name = string, index_name = string })), [])
        ai_bindings                          = optional(list(object({ name = string, project_id = string })), [])
        mtls_certificates                    = optional(list(object({ name = string, certificate_id = string })), [])
        browsers                             = optional(list(object({ name = string })), [])
      }))
      production = optional(object({
        compatibility_date                   = optional(string, "")
        compatibility_flags                  = optional(list(string), [])
        always_use_latest_compatibility_date = optional(bool, false)
        build_image_major_version            = optional(number, 0)
        fail_open                            = optional(bool, false)
        usage_model                          = optional(string, "")
        limits                               = optional(object({ cpu_ms = optional(number, 0) }))
        placement                            = optional(object({ mode = optional(string, "") }))
        vars                                 = optional(map(string), {})
        secrets                              = optional(list(object({ name = string, value = string })), [])
        kv_namespaces                        = optional(list(object({ name = string, namespace_id = string })), [])
        d1_databases                         = optional(list(object({ name = string, database_id = string })), [])
        r2_buckets                           = optional(list(object({ name = string, bucket_name = string, jurisdiction = optional(string, "") })), [])
        queue_producers                      = optional(list(object({ name = string, queue_name = string })), [])
        hyperdrive_bindings                  = optional(list(object({ name = string, config_id = string })), [])
        services                             = optional(list(object({ name = string, service = string, entrypoint = optional(string, ""), environment = optional(string, "") })), [])
        durable_object_namespaces            = optional(list(object({ name = string, namespace_id = string })), [])
        analytics_engine_datasets            = optional(list(object({ name = string, dataset = string })), [])
        vectorize_bindings                   = optional(list(object({ name = string, index_name = string })), [])
        ai_bindings                          = optional(list(object({ name = string, project_id = string })), [])
        mtls_certificates                    = optional(list(object({ name = string, certificate_id = string })), [])
        browsers                             = optional(list(object({ name = string })), [])
      }))
    }))

    domains = optional(list(string), [])
  })
}
