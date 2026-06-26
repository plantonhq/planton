# Cloudflare Load Balancer pool: an account-scoped group of origins, health-checked
# by a referenced monitor and selected by one or more zone-scoped load balancers.
resource "cloudflare_load_balancer_pool" "main" {
  account_id = var.spec.account_id
  name       = var.spec.name

  origins = local.origins

  monitor       = local.monitor != "" ? local.monitor : null
  check_regions = length(local.check_regions) > 0 ? local.check_regions : null

  description     = var.spec.description != "" ? var.spec.description : null
  enabled         = var.spec.enabled
  minimum_origins = try(var.spec.minimum_origins, 0) > 0 ? var.spec.minimum_origins : null
  latitude        = var.spec.latitude
  longitude       = var.spec.longitude

  load_shedding       = local.load_shedding
  origin_steering     = local.origin_steering
  notification_filter = local.notification_filter
}
