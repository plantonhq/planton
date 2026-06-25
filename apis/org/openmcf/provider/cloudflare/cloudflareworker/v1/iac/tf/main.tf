# Fetch the pre-built worker bundle from R2 when spec.r2_bundle is set.
data "aws_s3_object" "bundle" {
  count    = local.use_bundle ? 1 : 0
  provider = aws.r2
  bucket   = var.spec.r2_bundle.bucket
  key      = var.spec.r2_bundle.path
}

# The Worker script and all of its bindings.
resource "cloudflare_workers_script" "main" {
  account_id  = var.spec.account_id
  script_name = local.script_name

  content     = local.script_content
  main_module = var.spec.main_module

  compatibility_date  = local.compatibility_date
  compatibility_flags = length(var.spec.compatibility_flags) > 0 ? var.spec.compatibility_flags : null

  bindings = length(local.bindings) > 0 ? local.bindings : null

  observability  = local.observability
  placement      = local.placement
  limits         = local.limits
  logpush        = var.spec.logpush
  tail_consumers = length(local.tail_consumers) > 0 ? local.tail_consumers : null
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
resource "cloudflare_workers_custom_domain" "main" {
  for_each = local.custom_domains_map

  account_id  = var.spec.account_id
  hostname    = each.value.hostname
  service     = cloudflare_workers_script.main.script_name
  environment = "production"
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
