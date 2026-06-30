# main.tf

# Create a standalone OpenStack Neutron security group rule.
resource "openstack_networking_secgroup_rule_v2" "main" {
  security_group_id = local.security_group_id

  direction = var.spec.direction
  ethertype = var.spec.ethertype

  # Protocol (null means all protocols).
  protocol = var.spec.protocol != "" ? var.spec.protocol : null

  # Port range (null means all ports for the protocol).
  port_range_min = var.spec.port_range_min
  port_range_max = var.spec.port_range_max

  # Remote source -- mutually exclusive (enforced by proto CEL validation).
  remote_ip_prefix = var.spec.remote_ip_prefix != "" ? var.spec.remote_ip_prefix : null
  remote_group_id  = local.remote_group_id != "" ? local.remote_group_id : null

  # Per-rule description.
  description = var.spec.description != "" ? var.spec.description : null

  # Region override (optional).
  region = var.spec.region != "" ? var.spec.region : null
}
