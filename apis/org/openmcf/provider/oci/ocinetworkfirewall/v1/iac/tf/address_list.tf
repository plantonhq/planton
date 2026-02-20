resource "oci_network_firewall_network_firewall_policy_address_list" "this" {
  for_each = { for al in var.spec.policy.address_lists : al.name => al }

  network_firewall_policy_id = oci_network_firewall_network_firewall_policy.this.id
  name                       = each.value.name
  type                       = local.address_list_type_map[each.value.type]
  addresses                  = each.value.addresses

  description = each.value.description != "" ? each.value.description : null
}
