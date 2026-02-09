# main.tf

# Create the OpenStack Octavia pool member.
resource "openstack_lb_member_v2" "main" {
  name          = local.member_name
  pool_id       = local.pool_id
  address       = var.spec.address
  protocol_port = var.spec.protocol_port

  # Subnet ID (optional, for cross-subnet routing)
  subnet_id = local.subnet_id

  # Weight (optional, default 1 set by Octavia)
  weight = var.spec.weight != null ? var.spec.weight : null

  # Administrative state
  admin_state_up = var.spec.admin_state_up

  # Tags applied to the OpenStack resource
  tags = length(var.spec.tags) > 0 ? toset(var.spec.tags) : null

  # Region override (optional)
  region = var.spec.region != "" ? var.spec.region : null
}
