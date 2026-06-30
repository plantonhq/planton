resource "oci_network_firewall_network_firewall_policy_url_list" "this" {
  for_each = { for ul in var.spec.policy.url_lists : ul.name => ul }

  network_firewall_policy_id = oci_network_firewall_network_firewall_policy.this.id
  name                       = each.value.name

  dynamic "urls" {
    for_each = each.value.urls
    content {
      pattern = urls.value.pattern
      type    = "SIMPLE"
    }
  }

  description = each.value.description != "" ? each.value.description : null
}
