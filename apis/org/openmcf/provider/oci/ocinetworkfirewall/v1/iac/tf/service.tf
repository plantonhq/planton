resource "oci_network_firewall_network_firewall_policy_service" "this" {
  for_each = { for svc in var.spec.policy.services : svc.name => svc }

  network_firewall_policy_id = oci_network_firewall_network_firewall_policy.this.id
  name                       = each.value.name
  type                       = local.service_type_map[each.value.type]

  dynamic "port_ranges" {
    for_each = each.value.port_ranges
    content {
      minimum_port = port_ranges.value.minimum_port
      maximum_port = port_ranges.value.maximum_port
    }
  }

  description = each.value.description != "" ? each.value.description : null
}

resource "oci_network_firewall_network_firewall_policy_service_list" "this" {
  for_each = { for sl in var.spec.policy.service_lists : sl.name => sl }

  network_firewall_policy_id = oci_network_firewall_network_firewall_policy.this.id
  name                       = each.value.name
  services                   = each.value.services

  description = each.value.description != "" ? each.value.description : null

  depends_on = [oci_network_firewall_network_firewall_policy_service.this]
}
