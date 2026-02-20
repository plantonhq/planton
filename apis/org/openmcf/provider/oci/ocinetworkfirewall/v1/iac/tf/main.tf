resource "oci_network_firewall_network_firewall" "this" {
  compartment_id             = var.spec.compartment_id.value
  network_firewall_policy_id = oci_network_firewall_network_firewall_policy.this.id
  subnet_id                  = var.spec.subnet_id.value
  display_name               = local.display_name
  freeform_tags              = local.freeform_tags

  ipv4address         = var.spec.ipv4_address != "" ? var.spec.ipv4_address : null
  ipv6address         = var.spec.ipv6_address != "" ? var.spec.ipv6_address : null
  availability_domain = var.spec.availability_domain != "" ? var.spec.availability_domain : null
  shape               = var.spec.shape != "" ? var.spec.shape : null

  network_security_group_ids = length(var.spec.network_security_group_ids) > 0 ? [
    for n in var.spec.network_security_group_ids : n.value
  ] : null

  dynamic "nat_configuration" {
    for_each = var.spec.nat_configuration != null ? [var.spec.nat_configuration] : []
    content {
      must_enable_private_nat = nat_configuration.value.must_enable_private_nat
    }
  }

  depends_on = [
    oci_network_firewall_network_firewall_policy_security_rule.this,
  ]
}
