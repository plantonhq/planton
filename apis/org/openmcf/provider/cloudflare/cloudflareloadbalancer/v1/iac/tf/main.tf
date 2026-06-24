# The load balancer is zone-scoped, while the pool and monitor are account-scoped.
# Derive the account id from the zone so the spec only needs the zone reference.
data "cloudflare_zone" "this" {
  zone_id = local.zone_id
}

# Cloudflare Load Balancer Monitor (Health Check)
# This is an account-level resource that probes origins
resource "cloudflare_load_balancer_monitor" "health_check" {
  account_id     = data.cloudflare_zone.this.account.id
  type           = "https"
  method         = "GET"
  path           = local.health_probe_path
  expected_codes = "2xx"
  timeout        = 5
  interval       = 60 # 60 seconds minimum for Pro plan
  retries        = 2

  description = "Health check for ${local.resource_name}"
}

# Cloudflare Load Balancer Pool
# This is an account-level resource that groups origins
resource "cloudflare_load_balancer_pool" "main" {
  account_id = data.cloudflare_zone.this.account.id
  name       = "${local.resource_name}-pool"
  monitor    = cloudflare_load_balancer_monitor.health_check.id
  enabled    = true

  # One origin object per origin in the spec
  origins = [
    for o in local.origins : {
      name    = o.name
      address = o.address
      enabled = true
      weight  = coalesce(o.weight, 1)
    }
  ]

  description = "Pool for ${local.resource_name}"
}

# Cloudflare Load Balancer
# This is a zone-level resource that ties everything together
resource "cloudflare_load_balancer" "main" {
  zone_id = local.zone_id
  name    = var.spec.hostname

  # Pool configuration
  default_pools = [cloudflare_load_balancer_pool.main.id]
  fallback_pool = cloudflare_load_balancer_pool.main.id

  # Proxy configuration
  proxied = local.proxied

  # Traffic steering
  steering_policy = local.steering_policy

  # Session affinity
  session_affinity = local.session_affinity

  description = "Load balancer for ${var.spec.hostname}"

  # Wait for the pool to be created before creating the load balancer
  depends_on = [
    cloudflare_load_balancer_pool.main
  ]
}
