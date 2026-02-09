# main.tf

# Create the OpenStack Octavia health monitor.
# Note: Health monitors do NOT support tags in the Terraform provider.
resource "openstack_lb_monitor_v2" "main" {
  name       = local.monitor_name
  pool_id    = local.pool_id
  type       = var.spec.type
  delay      = var.spec.delay
  timeout    = var.spec.timeout
  max_retries = var.spec.max_retries

  # Max retries down (optional)
  max_retries_down = var.spec.max_retries_down != null ? var.spec.max_retries_down : null

  # URL path for HTTP/HTTPS monitors
  url_path = var.spec.url_path != "" ? var.spec.url_path : null

  # HTTP method for HTTP/HTTPS monitors
  http_method = var.spec.http_method != "" ? var.spec.http_method : null

  # Expected response codes for HTTP/HTTPS monitors
  expected_codes = var.spec.expected_codes != "" ? var.spec.expected_codes : null

  # Administrative state
  admin_state_up = var.spec.admin_state_up

  # Region override (optional)
  region = var.spec.region != "" ? var.spec.region : null
}
