# main.tf

# Create the OpenStack Octavia listener.
resource "openstack_lb_listener_v2" "main" {
  name            = local.listener_name
  loadbalancer_id = local.loadbalancer_id
  protocol        = var.spec.protocol
  protocol_port   = var.spec.protocol_port

  # Description (empty means unset)
  description = var.spec.description != "" ? var.spec.description : null

  # Connection limit (null means use Octavia default)
  connection_limit = var.spec.connection_limit

  # TLS container reference (required for TERMINATED_HTTPS)
  default_tls_container_ref = var.spec.default_tls_container_ref != "" ? var.spec.default_tls_container_ref : null

  # HTTP headers to insert before forwarding to backends
  insert_headers = length(var.spec.insert_headers) > 0 ? var.spec.insert_headers : null

  # Allowed CIDRs for access control
  allowed_cidrs = length(var.spec.allowed_cidrs) > 0 ? var.spec.allowed_cidrs : null

  # Administrative state
  admin_state_up = var.spec.admin_state_up

  # Tags applied to the OpenStack resource
  tags = length(var.spec.tags) > 0 ? toset(var.spec.tags) : null

  # Region override (optional)
  region = var.spec.region != "" ? var.spec.region : null
}
