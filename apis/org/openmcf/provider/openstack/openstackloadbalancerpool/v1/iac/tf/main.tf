# main.tf

# Create the OpenStack Octavia pool.
resource "openstack_lb_pool_v2" "main" {
  name        = local.pool_name
  listener_id = local.listener_id
  protocol    = var.spec.protocol
  lb_method   = var.spec.lb_method

  # Description (empty means unset)
  description = var.spec.description != "" ? var.spec.description : null

  # Administrative state
  admin_state_up = var.spec.admin_state_up

  # Session persistence (optional block)
  dynamic "persistence" {
    for_each = var.spec.persistence != null ? [var.spec.persistence] : []
    content {
      type        = persistence.value.type
      cookie_name = persistence.value.cookie_name != "" ? persistence.value.cookie_name : null
    }
  }

  # Tags applied to the OpenStack resource
  tags = length(var.spec.tags) > 0 ? toset(var.spec.tags) : null

  # Region override (optional)
  region = var.spec.region != "" ? var.spec.region : null
}
