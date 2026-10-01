locals {
  script_name = var.spec.worker_name

  # Script source: inline content, else the R2 bundle's bytes (decoded from
  # body_base64; a Worker script is UTF-8 text). A worker may instead be a pure
  # static site (assets only, no script) — in that case content and main_module
  # are null so the provider treats it as an assets-only Worker.
  use_bundle     = var.spec.r2_bundle != null
  has_script     = var.spec.content != "" || local.use_bundle
  script_content = local.has_script ? (var.spec.content != "" ? var.spec.content : base64decode(data.aws_s3_object.bundle[0].body_base64)) : null

  # The R2 S3-compatible endpoint the bundle read signs against: the connection's
  # override when it carries one, else the endpoint of the Worker's own account.
  r2_endpoint = coalesce(var.r2_endpoint, "https://${var.spec.account_id}.r2.cloudflarestorage.com")

  # body_part marks service-worker syntax and is mutually exclusive with
  # main_module (CEL). When body_part is set we omit main_module even if the
  # recommended default ("index.js") is present in variables.tf.
  main_module = local.has_script && try(var.spec.body_part, "") == "" ? (var.spec.main_module != "" ? var.spec.main_module : "index.js") : null
  body_part   = try(var.spec.body_part, "") != "" ? var.spec.body_part : null

  # Compatibility date defaults to today when unset.
  compatibility_date = var.spec.compatibility_date != "" ? var.spec.compatibility_date : formatdate("YYYY-MM-DD", timestamp())

  # Every flattened binding object carries the same attribute set (unused ones
  # null) so the provider's bindings list has a single, uniform object type.
  # Nested objects (outbound, simple) stay null on types that don't use them.
  null_attrs = {
    namespace_id                  = null
    bucket_name                   = null
    jurisdiction                  = null
    id                            = null
    text                          = null
    service                       = null
    environment                   = null
    entrypoint                    = null
    queue_name                    = null
    class_name                    = null
    script_name                   = null
    dataset                       = null
    index_name                    = null
    namespace                     = null
    outbound                      = null
    certificate_id                = null
    json                          = null
    pipeline                      = null
    simple                        = null
    secret_name                   = null
    store_id                      = null
    algorithm                     = null
    format                        = null
    usages                        = null
    key_base64                    = null
    key_jwk                       = null
    workflow_name                 = null
    part                          = null
    old_name                      = null
    version_id                    = null
    allowed_destination_addresses = null
    allowed_sender_addresses      = null
    destination_address           = null
    dispatch_namespace            = null
    service_id                    = null
    network_id                    = null
    tunnel_id                     = null
    instance_name                 = null
    database_id                   = null
    app_id                        = null
  }

  bindings = concat(
    [for k, v in var.spec.vars : merge(local.null_attrs, { name = k, type = "plain_text", text = v })],
    [for b in var.spec.secrets : merge(local.null_attrs, { name = b.name, type = "secret_text", text = b.value })],
    [for b in var.spec.kv_namespaces : merge(local.null_attrs, { name = b.name, type = "kv_namespace", namespace_id = b.namespace_id })],
    [for b in var.spec.r2_buckets : merge(local.null_attrs, { name = b.name, type = "r2_bucket", bucket_name = b.bucket_name, jurisdiction = b.jurisdiction != "" ? b.jurisdiction : null })],
    [for b in var.spec.d1_databases : merge(local.null_attrs, { name = b.name, type = "d1", id = b.database_id })],
    [for b in var.spec.hyperdrive_configs : merge(local.null_attrs, { name = b.name, type = "hyperdrive", id = b.config_id })],
    [for b in var.spec.services : merge(local.null_attrs, { name = b.name, type = "service", service = b.service, environment = b.environment != "" ? b.environment : null, entrypoint = b.entrypoint != "" ? b.entrypoint : null })],
    [for b in var.spec.queues : merge(local.null_attrs, { name = b.name, type = "queue", queue_name = b.queue_name })],
    [for b in var.spec.durable_objects : merge(local.null_attrs, { name = b.name, type = "durable_object_namespace", class_name = b.class_name, script_name = try(b.script_name, "") != "" ? b.script_name : null, environment = try(b.environment, "") != "" ? b.environment : null, namespace_id = try(b.namespace_id, "") != "" ? b.namespace_id : null, dispatch_namespace = try(b.dispatch_namespace, "") != "" ? b.dispatch_namespace : null })],
    [for b in var.spec.analytics_engine_datasets : merge(local.null_attrs, { name = b.name, type = "analytics_engine", dataset = b.dataset })],
    [for b in var.spec.vectorize_indexes : merge(local.null_attrs, { name = b.name, type = "vectorize", index_name = b.index_name })],
    [for b in var.spec.ai : merge(local.null_attrs, { name = b.name, type = "ai" })],
    [for b in var.spec.version_metadata : merge(local.null_attrs, { name = b.name, type = "version_metadata" })],
    [for b in try(var.spec.mtls_certificates, []) : merge(local.null_attrs, { name = b.name, type = "mtls_certificate", certificate_id = b.certificate_id })],
    [for b in try(var.spec.dispatch_namespaces, []) : merge(local.null_attrs, { name = b.name, type = "dispatch_namespace", namespace = b.namespace, outbound = try(b.outbound, null) != null ? { params = try(b.outbound.params, []), worker = try(b.outbound.worker, null) != null ? { service = try(b.outbound.worker.service, null), environment = try(b.outbound.worker.environment, "") != "" ? b.outbound.worker.environment : null } : null } : null })],
    [for b in try(var.spec.rate_limits, []) : merge(local.null_attrs, { name = b.name, type = "ratelimit", namespace = b.namespace, simple = { limit = b.simple.limit, period = b.simple.period, mitigation_timeout = try(b.simple.mitigation_timeout, 0) > 0 ? b.simple.mitigation_timeout : null } })],
    [for b in try(var.spec.send_email, []) : merge(local.null_attrs, { name = b.name, type = "send_email", destination_address = try(b.destination_address, "") != "" ? b.destination_address : null, allowed_destination_addresses = length(try(b.allowed_destination_addresses, [])) > 0 ? b.allowed_destination_addresses : null, allowed_sender_addresses = length(try(b.allowed_sender_addresses, [])) > 0 ? b.allowed_sender_addresses : null })],
    [for b in try(var.spec.secrets_store_secrets, []) : merge(local.null_attrs, { name = b.name, type = "secrets_store_secret", store_id = b.store_id, secret_name = b.secret_name })],
    [for b in try(var.spec.secret_keys, []) : merge(local.null_attrs, { name = b.name, type = "secret_key", algorithm = b.algorithm, format = b.format, usages = length(try(b.usages, [])) > 0 ? b.usages : null, key_base64 = try(b.key_base64, "") != "" ? b.key_base64 : null, key_jwk = try(b.key_jwk, "") != "" ? b.key_jwk : null })],
    [for b in try(var.spec.workflows, []) : merge(local.null_attrs, { name = b.name, type = "workflow", workflow_name = b.workflow_name })],
    [for b in try(var.spec.pipelines, []) : merge(local.null_attrs, { name = b.name, type = "pipelines", pipeline = b.pipeline })],
    [for b in try(var.spec.json_bindings, []) : merge(local.null_attrs, { name = b.name, type = "json", json = b.json })],
    [for b in try(var.spec.inherit_bindings, []) : merge(local.null_attrs, { name = b.name, type = "inherit", old_name = try(b.old_name, "") != "" ? b.old_name : null, version_id = try(b.version_id, "") != "" ? b.version_id : null })],
    [for b in try(var.spec.data_blobs, []) : merge(local.null_attrs, { name = b.name, type = "data_blob", part = b.part })],
    [for b in try(var.spec.text_blobs, []) : merge(local.null_attrs, { name = b.name, type = "text_blob", part = b.part })],
    [for b in try(var.spec.browsers, []) : merge(local.null_attrs, { name = b.name, type = "browser" })],
    [for b in try(var.spec.ai_search, []) : merge(local.null_attrs, { name = b.name, type = "ai_search", instance_name = b.instance_name, namespace = try(b.namespace, "") != "" ? b.namespace : null, app_id = try(b.app_id, "") != "" ? b.app_id : null })],
    [for b in try(var.spec.ai_search_namespaces, []) : merge(local.null_attrs, { name = b.name, type = "ai_search_namespace", namespace = b.namespace })],
    [for b in try(var.spec.images, []) : merge(local.null_attrs, { name = b.name, type = "images" })],
    [for b in try(var.spec.media, []) : merge(local.null_attrs, { name = b.name, type = "media" })],
    [for b in try(var.spec.wasm_modules, []) : merge(local.null_attrs, { name = b.name, type = "wasm_module", part = b.part })],
    [for b in try(var.spec.vpc_services, []) : merge(local.null_attrs, { name = b.name, type = "vpc_service", service_id = b.service_id })],
    [for b in try(var.spec.vpc_networks, []) : merge(local.null_attrs, { name = b.name, type = "vpc_network", network_id = try(b.network_id, "") != "" ? b.network_id : null, tunnel_id = try(b.tunnel_id, "") != "" ? b.tunnel_id : null })],
    [for b in try(var.spec.tail_consumer_bindings, []) : merge(local.null_attrs, { name = b.name, type = "tail_consumer", service = b.service })],
    # Assets binding (env.<NAME>) when a full-stack worker wants programmatic access.
    (try(var.spec.assets.binding_name, "") != "") ? [merge(local.null_attrs, { name = var.spec.assets.binding_name, type = "assets" })] : [],
  )

  # Workers Static Assets configuration. The provider's run_worker_first is a
  # DYNAMIC field that accepts either a bool (apply to all paths) or a list of
  # path rules. HCL conditionals must unify to one type, so we cannot write a
  # `rules : bool` ternary directly — instead we encode the chosen value to JSON
  # (always a string) and decode it, which defers the bool-or-list typing to
  # runtime and lets us feed the dynamic attribute. null means "not configured".
  assets = try(var.spec.assets, null) != null ? {
    directory = var.spec.assets.directory
    config = try(var.spec.assets.config, null) != null ? {
      html_handling      = var.spec.assets.config.html_handling != "" ? var.spec.assets.config.html_handling : null
      not_found_handling = var.spec.assets.config.not_found_handling != "" ? var.spec.assets.config.not_found_handling : null
      headers            = var.spec.assets.config.headers != "" ? var.spec.assets.config.headers : null
      redirects          = var.spec.assets.config.redirects != "" ? var.spec.assets.config.redirects : null
      run_worker_first = jsondecode(
        length(var.spec.assets.config.run_worker_first_rules) > 0
        ? jsonencode(var.spec.assets.config.run_worker_first_rules)
        : (var.spec.assets.config.run_worker_first ? "true" : "null")
      )
    } : null
  } : null

  observability = try(var.spec.observability, null) != null ? {
    enabled            = try(var.spec.observability.enabled, false)
    head_sampling_rate = try(var.spec.observability.head_sampling_rate, 0) > 0 ? var.spec.observability.head_sampling_rate : null
    logs = try(var.spec.observability.logs, null) != null ? {
      enabled            = var.spec.observability.logs.enabled
      invocation_logs    = var.spec.observability.logs.invocation_logs
      destinations       = length(try(var.spec.observability.logs.destinations, [])) > 0 ? var.spec.observability.logs.destinations : null
      head_sampling_rate = try(var.spec.observability.logs.head_sampling_rate, 0) > 0 ? var.spec.observability.logs.head_sampling_rate : null
      persist            = try(var.spec.observability.logs.persist, false)
    } : null
    traces = try(var.spec.observability.traces, null) != null ? {
      destinations       = length(try(var.spec.observability.traces.destinations, [])) > 0 ? var.spec.observability.traces.destinations : null
      enabled            = try(var.spec.observability.traces.enabled, false)
      head_sampling_rate = try(var.spec.observability.traces.head_sampling_rate, 0) > 0 ? var.spec.observability.traces.head_sampling_rate : null
      persist            = try(var.spec.observability.traces.persist, false)
      propagation_policy = try(var.spec.observability.traces.propagation_policy, "") != "" ? var.spec.observability.traces.propagation_policy : null
    } : null
  } : null

  placement = (try(var.spec.placement.mode, "") != "") ? { mode = var.spec.placement.mode } : null

  limits = (try(var.spec.limits.cpu_ms, 0) > 0 || try(var.spec.limits.subrequests, 0) > 0) ? {
    cpu_ms      = try(var.spec.limits.cpu_ms, 0) > 0 ? var.spec.limits.cpu_ms : null
    subrequests = try(var.spec.limits.subrequests, 0) > 0 ? var.spec.limits.subrequests : null
  } : null

  migrations = try(var.spec.migrations, null) != null ? {
    deleted_classes    = length(try(var.spec.migrations.deleted_classes, [])) > 0 ? var.spec.migrations.deleted_classes : null
    new_classes        = length(try(var.spec.migrations.new_classes, [])) > 0 ? var.spec.migrations.new_classes : null
    new_sqlite_classes = length(try(var.spec.migrations.new_sqlite_classes, [])) > 0 ? var.spec.migrations.new_sqlite_classes : null
    new_tag            = try(var.spec.migrations.new_tag, "") != "" ? var.spec.migrations.new_tag : null
    old_tag            = try(var.spec.migrations.old_tag, "") != "" ? var.spec.migrations.old_tag : null
    renamed_classes    = length(try(var.spec.migrations.renamed_classes, [])) > 0 ? var.spec.migrations.renamed_classes : null
    transferred_classes = length(try(var.spec.migrations.transferred_classes, [])) > 0 ? [for t in var.spec.migrations.transferred_classes : {
      from        = try(t.from, "") != "" ? t.from : null
      from_script = try(t.from_script, "") != "" ? t.from_script : null
      to          = try(t.to, "") != "" ? t.to : null
    }] : null
    steps = length(try(var.spec.migrations.steps, [])) > 0 ? [for s in var.spec.migrations.steps : {
      deleted_classes    = length(try(s.deleted_classes, [])) > 0 ? s.deleted_classes : null
      new_classes        = length(try(s.new_classes, [])) > 0 ? s.new_classes : null
      new_sqlite_classes = length(try(s.new_sqlite_classes, [])) > 0 ? s.new_sqlite_classes : null
      renamed_classes    = length(try(s.renamed_classes, [])) > 0 ? s.renamed_classes : null
      transferred_classes = length(try(s.transferred_classes, [])) > 0 ? [for t in s.transferred_classes : {
        from        = try(t.from, "") != "" ? t.from : null
        from_script = try(t.from_script, "") != "" ? t.from_script : null
        to          = try(t.to, "") != "" ? t.to : null
      }] : null
    }] : null
  } : null

  cache_options = try(var.spec.cache_options, null) != null ? {
    enabled             = try(var.spec.cache_options.enabled, false)
    cross_version_cache = try(var.spec.cache_options.cross_version_cache, false)
  } : null

  exports = try(var.spec.exports, null) != null && length(try(var.spec.exports, {})) > 0 ? {
    for k, v in var.spec.exports : k => {
      type  = v.type
      cache = try(v.cache, null) != null ? { enabled = v.cache.enabled } : null
    }
  } : null

  package_dependencies = length(try(var.spec.package_dependencies, [])) > 0 ? [for p in var.spec.package_dependencies : {
    name                 = p.name
    installed_version    = p.installed_version
    package_json_version = p.package_json_version
  }] : null

  annotations = try(var.spec.annotations, null) != null ? {
    workers_message = try(var.spec.annotations.workers_message, "") != "" ? var.spec.annotations.workers_message : null
    workers_tag     = try(var.spec.annotations.workers_tag, "") != "" ? var.spec.annotations.workers_tag : null
  } : null

  workers_dev_enabled = try(var.spec.workers_dev.enabled, false)

  custom_domains_map = { for cd in var.spec.custom_domains : cd.hostname => cd }
  routes_map         = { for idx, r in var.spec.routes : tostring(idx) => r }

  tail_consumers = [for t in var.spec.tail_consumers : {
    service     = t.service
    environment = t.environment != "" ? t.environment : null
    namespace   = t.namespace != "" ? t.namespace : null
  }]
}
