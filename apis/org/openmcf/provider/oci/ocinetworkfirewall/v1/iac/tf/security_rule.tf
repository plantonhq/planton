resource "oci_network_firewall_network_firewall_policy_security_rule" "this" {
  for_each = { for idx, rule in var.spec.policy.security_rules : rule.name => merge(rule, { priority = idx + 1 }) }

  network_firewall_policy_id = oci_network_firewall_network_firewall_policy.this.id
  name                       = each.value.name
  action                     = local.action_map[each.value.action]
  priority_order             = tostring(each.value.priority)

  condition {
    source_address      = length(each.value.condition.source_addresses) > 0 ? each.value.condition.source_addresses : []
    destination_address = length(each.value.condition.destination_addresses) > 0 ? each.value.condition.destination_addresses : []
    service             = length(each.value.condition.services) > 0 ? each.value.condition.services : []
    url                 = length(each.value.condition.urls) > 0 ? each.value.condition.urls : []
  }

  inspection  = each.value.inspection != "" ? local.inspection_map[each.value.inspection] : null
  description = each.value.description != "" ? each.value.description : null

  depends_on = [
    oci_network_firewall_network_firewall_policy_address_list.this,
    oci_network_firewall_network_firewall_policy_service.this,
    oci_network_firewall_network_firewall_policy_service_list.this,
    oci_network_firewall_network_firewall_policy_url_list.this,
  ]
}
