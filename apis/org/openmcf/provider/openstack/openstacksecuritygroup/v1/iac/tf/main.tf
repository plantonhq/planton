# main.tf

# Create the OpenStack Neutron security group.
resource "openstack_networking_secgroup_v2" "main" {
  name        = local.sg_name
  description = var.spec.description != "" ? var.spec.description : null

  # Delete default egress rules after creation (create-time only).
  # When true, the default allow-all-egress IPv4 and IPv6 rules are removed.
  delete_default_rules = var.spec.delete_default_rules

  # Stateful/stateless mode (null lets OpenStack decide).
  stateful = var.spec.stateful

  # Tags applied to the OpenStack resource.
  tags = length(var.spec.tags) > 0 ? toset(var.spec.tags) : null

  # Region override (optional).
  region = var.spec.region != "" ? var.spec.region : null
}

# Create inline security group rules.
# Each rule is keyed by the user-provided `key` field for stable state management.
# Adding, removing, or reordering rules only affects the specific rule changed.
resource "openstack_networking_secgroup_rule_v2" "rules" {
  for_each = local.rules_map

  security_group_id = openstack_networking_secgroup_v2.main.id

  direction = each.value.direction
  ethertype = each.value.ethertype

  # Protocol (null means all protocols).
  protocol = each.value.protocol != "" ? each.value.protocol : null

  # Port range (null means all ports for the protocol).
  port_range_min = each.value.port_range_min
  port_range_max = each.value.port_range_max

  # Remote source -- mutually exclusive (enforced by proto CEL validation).
  remote_ip_prefix = each.value.remote_ip_prefix != "" ? each.value.remote_ip_prefix : null
  remote_group_id  = each.value.remote_group_id != "" ? each.value.remote_group_id : null

  # Per-rule description.
  description = each.value.description != "" ? each.value.description : null

  # Region matches the security group.
  region = var.spec.region != "" ? var.spec.region : null
}
