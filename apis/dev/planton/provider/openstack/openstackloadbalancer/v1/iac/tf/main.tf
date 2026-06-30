# main.tf

# Create the OpenStack Octavia load balancer.
resource "openstack_lb_loadbalancer_v2" "main" {
  name          = local.lb_name
  vip_subnet_id = local.vip_subnet_id

  # VIP address (empty means auto-allocate)
  vip_address = var.spec.vip_address != "" ? var.spec.vip_address : null

  # Description (empty means unset)
  description = var.spec.description != "" ? var.spec.description : null

  # Administrative state
  admin_state_up = var.spec.admin_state_up

  # Octavia flavor
  flavor_id = var.spec.flavor_id != "" ? var.spec.flavor_id : null

  # Tags applied to the OpenStack resource
  tags = length(var.spec.tags) > 0 ? toset(var.spec.tags) : null

  # Region override (optional)
  region = var.spec.region != "" ? var.spec.region : null
}
