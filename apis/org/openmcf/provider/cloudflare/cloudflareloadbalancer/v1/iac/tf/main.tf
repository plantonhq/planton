# Cloudflare Load Balancer: a zone-scoped resource that attaches a DNS hostname to
# account-scoped pools (CloudflareLoadBalancerPool) and steers traffic across them.
# Pools and their monitors are separate, reusable resources referenced by ID/ref.
resource "cloudflare_load_balancer" "main" {
  zone_id = local.zone_id
  name    = var.spec.hostname

  default_pools = var.spec.default_pools
  fallback_pool = var.spec.fallback_pool

  description = var.spec.description != "" ? var.spec.description : null

  proxied = var.spec.proxied
  enabled = var.spec.enabled

  steering_policy  = local.steering_policy
  session_affinity = local.session_affinity

  ttl                  = var.spec.ttl > 0 ? var.spec.ttl : null
  session_affinity_ttl = var.spec.session_affinity_ttl > 0 ? var.spec.session_affinity_ttl : null

  session_affinity_attributes = local.saa

  region_pools  = length(local.region_pools) > 0 ? local.region_pools : null
  country_pools = length(local.country_pools) > 0 ? local.country_pools : null
  pop_pools     = length(local.pop_pools) > 0 ? local.pop_pools : null

  adaptive_routing  = local.adaptive_routing
  location_strategy = local.location_strategy
  random_steering   = local.random_steering
}
