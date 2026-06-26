# Cloudflare Load Balancer monitor: an account-scoped health check that pools
# reference to probe their origins.
resource "cloudflare_load_balancer_monitor" "main" {
  account_id = var.spec.account_id
  type       = local.monitor_type

  description = var.spec.description != "" ? var.spec.description : null
  path        = var.spec.path != "" ? var.spec.path : null
  method      = var.spec.method != "" ? var.spec.method : null
  probe_zone  = var.spec.probe_zone != "" ? var.spec.probe_zone : null

  expected_codes = var.spec.expected_codes != "" ? var.spec.expected_codes : null
  expected_body  = var.spec.expected_body != "" ? var.spec.expected_body : null

  # 0 means "use the Cloudflare default" for these tuning knobs.
  port             = var.spec.port > 0 ? var.spec.port : null
  interval         = var.spec.interval > 0 ? var.spec.interval : null
  timeout          = var.spec.timeout > 0 ? var.spec.timeout : null
  retries          = var.spec.retries > 0 ? var.spec.retries : null
  consecutive_up   = var.spec.consecutive_up > 0 ? var.spec.consecutive_up : null
  consecutive_down = var.spec.consecutive_down > 0 ? var.spec.consecutive_down : null

  follow_redirects = var.spec.follow_redirects
  allow_insecure   = var.spec.allow_insecure

  header = length(local.headers) > 0 ? local.headers : null
}
