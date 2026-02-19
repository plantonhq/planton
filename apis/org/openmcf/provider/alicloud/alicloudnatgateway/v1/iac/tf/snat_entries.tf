resource "alicloud_snat_entry" "entries" {
  for_each = local.snat_entries_map

  snat_table_id    = alicloud_nat_gateway.main.snat_table_ids
  snat_ip          = data.alicloud_eip_addresses.nat.addresses[0].ip_address
  source_vswitch_id = each.value.source_vswitch_id != "" ? each.value.source_vswitch_id : null
  source_cidr      = each.value.source_cidr != "" ? each.value.source_cidr : null
  snat_entry_name  = each.key

  depends_on = [alicloud_eip_association.nat]
}
