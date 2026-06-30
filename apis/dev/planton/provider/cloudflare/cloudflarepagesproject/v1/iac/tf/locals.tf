locals {
  build_config = try(var.spec.build_config, null) == null ? null : {
    build_command       = var.spec.build_config.build_command != "" ? var.spec.build_config.build_command : null
    destination_dir     = var.spec.build_config.destination_dir != "" ? var.spec.build_config.destination_dir : null
    root_dir            = var.spec.build_config.root_dir != "" ? var.spec.build_config.root_dir : null
    build_caching       = var.spec.build_config.build_caching
    web_analytics_tag   = var.spec.build_config.web_analytics_tag != "" ? var.spec.build_config.web_analytics_tag : null
    web_analytics_token = var.spec.build_config.web_analytics_token != "" ? var.spec.build_config.web_analytics_token : null
  }

  source = try(var.spec.source, null) == null ? null : {
    type = var.spec.source.type
    config = {
      owner                          = var.spec.source.config.owner != "" ? var.spec.source.config.owner : null
      repo_name                      = var.spec.source.config.repo_name != "" ? var.spec.source.config.repo_name : null
      production_branch              = var.spec.source.config.production_branch != "" ? var.spec.source.config.production_branch : null
      pr_comments_enabled            = var.spec.source.config.pr_comments_enabled
      deployments_enabled            = var.spec.source.config.deployments_enabled
      production_deployments_enabled = var.spec.source.config.production_deployments_enabled
      preview_deployment_setting     = var.spec.source.config.preview_deployment_setting != "" ? var.spec.source.config.preview_deployment_setting : null
      preview_branch_includes        = length(var.spec.source.config.preview_branch_includes) > 0 ? var.spec.source.config.preview_branch_includes : null
      preview_branch_excludes        = length(var.spec.source.config.preview_branch_excludes) > 0 ? var.spec.source.config.preview_branch_excludes : null
      path_includes                  = length(var.spec.source.config.path_includes) > 0 ? var.spec.source.config.path_includes : null
      path_excludes                  = length(var.spec.source.config.path_excludes) > 0 ? var.spec.source.config.path_excludes : null
    }
  }

  # Cloudflare treats preview and production as a paired configuration and rejects
  # a project whose environments are configured inconsistently (e.g. fail_open must
  # match). When only one environment is supplied we therefore mirror it to both,
  # so a single config "just works"; supply both explicitly to differ them.
  dc_preview_raw    = try(var.spec.deployment_configs.preview, null)
  dc_production_raw = try(var.spec.deployment_configs.production, null)
  dc_preview_src    = local.dc_preview_raw != null ? local.dc_preview_raw : local.dc_production_raw
  dc_production_src = local.dc_production_raw != null ? local.dc_production_raw : local.dc_preview_raw

  # preview and production share one shape; transform both with a single
  # expression so the two are always normalized identically.
  dc_in = {
    preview    = local.dc_preview_src
    production = local.dc_production_src
  }

  dc_built = {
    for env, c in local.dc_in : env => c == null ? null : {
      compatibility_date                   = c.compatibility_date != "" ? c.compatibility_date : null
      compatibility_flags                  = length(c.compatibility_flags) > 0 ? c.compatibility_flags : null
      # fail_open / always_use_latest_compatibility_date are omitted when false so
      # an unconfigured environment keeps the server default — Cloudflare rejects a
      # project whose fail_open differs between preview and production.
      always_use_latest_compatibility_date = c.always_use_latest_compatibility_date ? true : null
      build_image_major_version            = c.build_image_major_version > 0 ? c.build_image_major_version : null
      fail_open                            = c.fail_open ? true : null
      usage_model                          = c.usage_model != "" ? c.usage_model : null
      limits                               = try(c.limits.cpu_ms, 0) > 0 ? { cpu_ms = c.limits.cpu_ms } : null
      placement                            = try(c.placement.mode, "") != "" ? { mode = c.placement.mode } : null
      # Empty collections are sent as null (omitted), not {}: the provider
      # normalizes an empty map to null and would otherwise flag an inconsistent
      # apply result.
      env_vars = (length(c.vars) + length(c.secrets)) > 0 ? merge(
        { for k, v in c.vars : k => { type = "plain_text", value = v } },
        { for s in c.secrets : s.name => { type = "secret_text", value = s.value } }
      ) : null
      kv_namespaces             = length(c.kv_namespaces) > 0 ? { for b in c.kv_namespaces : b.name => { namespace_id = b.namespace_id } } : null
      d1_databases              = length(c.d1_databases) > 0 ? { for b in c.d1_databases : b.name => { id = b.database_id } } : null
      r2_buckets                = length(c.r2_buckets) > 0 ? { for b in c.r2_buckets : b.name => { name = b.bucket_name, jurisdiction = b.jurisdiction != "" ? b.jurisdiction : null } } : null
      queue_producers           = length(c.queue_producers) > 0 ? { for b in c.queue_producers : b.name => { name = b.queue_name } } : null
      hyperdrive_bindings       = length(c.hyperdrive_bindings) > 0 ? { for b in c.hyperdrive_bindings : b.name => { id = b.config_id } } : null
      services                  = length(c.services) > 0 ? { for b in c.services : b.name => { service = b.service, entrypoint = b.entrypoint != "" ? b.entrypoint : null, environment = b.environment != "" ? b.environment : null } } : null
      durable_object_namespaces = length(c.durable_object_namespaces) > 0 ? { for b in c.durable_object_namespaces : b.name => { namespace_id = b.namespace_id } } : null
      analytics_engine_datasets = length(c.analytics_engine_datasets) > 0 ? { for b in c.analytics_engine_datasets : b.name => { dataset = b.dataset } } : null
      vectorize_bindings        = length(c.vectorize_bindings) > 0 ? { for b in c.vectorize_bindings : b.name => { index_name = b.index_name } } : null
      ai_bindings               = length(c.ai_bindings) > 0 ? { for b in c.ai_bindings : b.name => { project_id = b.project_id } } : null
      mtls_certificates         = length(c.mtls_certificates) > 0 ? { for b in c.mtls_certificates : b.name => { certificate_id = b.certificate_id } } : null
      browsers                  = length(c.browsers) > 0 ? { for b in c.browsers : b.name => {} } : null
    }
  }

  deployment_configs = (local.dc_preview_src == null && local.dc_production_src == null) ? null : {
    preview    = local.dc_built.preview
    production = local.dc_built.production
  }

  domains_map = { for d in var.spec.domains : d => d }
}
