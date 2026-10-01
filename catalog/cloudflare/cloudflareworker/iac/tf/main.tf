# Fetch the pre-built worker bundle from R2 when spec.r2_bundle is set.
#
# download_body returns the object's bytes as body_base64 whatever its Content-Type.
# The plain body attribute is filled only for a short list of "readable" types, which
# excludes application/javascript -- the type many upload tools give a .js file -- so a
# bundle read through it can come back empty. The postcondition refuses an empty
# bundle here, naming the object, instead of letting Cloudflare receive a Worker with
# no code.
data "aws_s3_object" "bundle" {
  count         = local.use_bundle ? 1 : 0
  provider      = aws.r2
  bucket        = var.spec.r2_bundle.bucket
  key           = var.spec.r2_bundle.path
  download_body = true

  lifecycle {
    postcondition {
      condition     = try(self.body_base64, "") != ""
      error_message = "The Worker bundle r2://${var.spec.r2_bundle.bucket}/${var.spec.r2_bundle.path} is empty. Upload the built script to that key, or point spec.r2_bundle.path at the object that holds it."
    }
  }
}

# The Worker script and all of its bindings.
resource "cloudflare_workers_script" "main" {
  account_id  = var.spec.account_id
  script_name = local.script_name

  content      = local.script_content
  main_module  = local.main_module
  body_part    = local.body_part
  content_type = try(var.spec.content_type, "") != "" ? var.spec.content_type : null

  compatibility_date  = local.compatibility_date
  compatibility_flags = length(var.spec.compatibility_flags) > 0 ? var.spec.compatibility_flags : null

  bindings = length(local.bindings) > 0 ? local.bindings : null

  # Workers Static Assets (built site directory served from the edge).
  assets = local.assets

  keep_assets   = try(var.spec.keep_assets, false) ? true : null
  keep_bindings = length(try(var.spec.keep_bindings, [])) > 0 ? var.spec.keep_bindings : null
  usage_model   = try(var.spec.usage_model, "") != "" ? var.spec.usage_model : null

  migrations     = local.migrations
  observability  = local.observability
  placement      = local.placement
  limits         = local.limits
  logpush        = var.spec.logpush
  tail_consumers = length(local.tail_consumers) > 0 ? local.tail_consumers : null

  # Pulumi SDK v6.17.0 has no matching inputs for these three — tofu honors
  # them; Pulumi logs a PARITY-EXCEPTION and skips them.
  cache_options        = local.cache_options
  exports              = local.exports
  package_dependencies = local.package_dependencies
  annotations          = local.annotations
}

# workers.dev subdomain exposure.
resource "cloudflare_workers_script_subdomain" "main" {
  count = local.workers_dev_enabled ? 1 : 0

  account_id       = var.spec.account_id
  script_name      = cloudflare_workers_script.main.script_name
  enabled          = true
  previews_enabled = try(var.spec.workers_dev.previews_enabled, false)
}

# Managed custom domains routed directly to the Worker.
# environment is deprecated on workers_custom_domain and is omitted — Cloudflare
# defaults the hostname to the production environment.
resource "cloudflare_workers_custom_domain" "main" {
  for_each = local.custom_domains_map

  account_id = var.spec.account_id
  hostname   = each.value.hostname
  service    = cloudflare_workers_script.main.script_name
  # Zone is optional — Cloudflare infers it from the hostname when omitted.
  zone_id = each.value.zone_id != "" ? each.value.zone_id : null
}

# Pattern-based routes mapping zone requests to the Worker.
resource "cloudflare_workers_route" "main" {
  for_each = local.routes_map

  zone_id = each.value.zone_id
  pattern = each.value.pattern
  script  = cloudflare_workers_script.main.script_name
}

# Cron-triggered invocations of the Worker's scheduled handler.
resource "cloudflare_workers_cron_trigger" "main" {
  count = length(var.spec.schedules) > 0 ? 1 : 0

  account_id  = var.spec.account_id
  script_name = cloudflare_workers_script.main.script_name
  schedules   = [for s in var.spec.schedules : { cron = s }]
}
