locals {
  script_name = var.spec.worker_name

  # Script source: inline content, else the R2 bundle body.
  use_bundle     = var.spec.r2_bundle != null
  script_content = var.spec.content != "" ? var.spec.content : (local.use_bundle ? data.aws_s3_object.bundle[0].body : "")

  # Compatibility date defaults to today when unset.
  compatibility_date = var.spec.compatibility_date != "" ? var.spec.compatibility_date : formatdate("YYYY-MM-DD", timestamp())

  # Every flattened binding object carries the same attribute set (unused ones
  # null) so the provider's bindings list has a single, uniform object type.
  null_attrs = {
    namespace_id = null
    bucket_name  = null
    jurisdiction = null
    id           = null
    text         = null
    service      = null
    environment  = null
    queue_name   = null
    class_name   = null
    script_name  = null
    dataset      = null
    index_name   = null
  }

  bindings = concat(
    [for k, v in var.spec.vars : merge(local.null_attrs, { name = k, type = "plain_text", text = v })],
    [for b in var.spec.secrets : merge(local.null_attrs, { name = b.name, type = "secret_text", text = b.value })],
    [for b in var.spec.kv_namespaces : merge(local.null_attrs, { name = b.name, type = "kv_namespace", namespace_id = b.namespace_id })],
    [for b in var.spec.r2_buckets : merge(local.null_attrs, { name = b.name, type = "r2_bucket", bucket_name = b.bucket_name, jurisdiction = b.jurisdiction != "" ? b.jurisdiction : null })],
    [for b in var.spec.d1_databases : merge(local.null_attrs, { name = b.name, type = "d1", id = b.database_id })],
    [for b in var.spec.hyperdrive_configs : merge(local.null_attrs, { name = b.name, type = "hyperdrive", id = b.config_id })],
    [for b in var.spec.services : merge(local.null_attrs, { name = b.name, type = "service", service = b.service, environment = b.environment != "" ? b.environment : null })],
    [for b in var.spec.queues : merge(local.null_attrs, { name = b.name, type = "queue", queue_name = b.queue_name })],
    [for b in var.spec.durable_objects : merge(local.null_attrs, { name = b.name, type = "durable_object_namespace", class_name = b.class_name, script_name = b.script_name != "" ? b.script_name : null, environment = b.environment != "" ? b.environment : null })],
    [for b in var.spec.analytics_engine_datasets : merge(local.null_attrs, { name = b.name, type = "analytics_engine", dataset = b.dataset })],
    [for b in var.spec.vectorize_indexes : merge(local.null_attrs, { name = b.name, type = "vectorize", index_name = b.index_name })],
    [for b in var.spec.ai : merge(local.null_attrs, { name = b.name, type = "ai" })],
    [for b in var.spec.version_metadata : merge(local.null_attrs, { name = b.name, type = "version_metadata" })],
  )

  observability = try(var.spec.observability, null) != null ? {
    enabled            = try(var.spec.observability.enabled, false)
    head_sampling_rate = try(var.spec.observability.head_sampling_rate, 0) > 0 ? var.spec.observability.head_sampling_rate : null
  } : null

  placement = (try(var.spec.placement.mode, "") != "") ? { mode = var.spec.placement.mode } : null

  limits = (try(var.spec.limits.cpu_ms, 0) > 0) ? {
    cpu_ms = var.spec.limits.cpu_ms
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
