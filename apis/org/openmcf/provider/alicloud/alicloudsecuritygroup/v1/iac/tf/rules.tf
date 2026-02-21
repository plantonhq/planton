resource "alicloud_security_group_rule" "rules" {
  for_each = local.rules_map

  security_group_id         = alicloud_security_group.main.id
  type                      = each.value.type
  ip_protocol               = each.value.ip_protocol
  port_range                = each.value.port_range
  nic_type                  = "intranet"
  cidr_ip                   = each.value.cidr_ip != "" ? each.value.cidr_ip : null
  source_security_group_id  = each.value.source_security_group_id != "" ? each.value.source_security_group_id : null
  priority                  = each.value.priority
  policy                    = each.value.policy
  description               = each.value.description != "" ? each.value.description : null
}
