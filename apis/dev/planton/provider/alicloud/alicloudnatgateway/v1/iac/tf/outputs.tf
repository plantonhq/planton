output "nat_gateway_id" {
  description = "The NAT Gateway ID"
  value       = alicloud_nat_gateway.main.id
}

output "nat_gateway_name" {
  description = "The NAT Gateway name"
  value       = alicloud_nat_gateway.main.nat_gateway_name
}

output "snat_table_id" {
  description = "The SNAT table ID"
  value       = alicloud_nat_gateway.main.snat_table_ids
}

output "forward_table_id" {
  description = "The forward (DNAT) table ID"
  value       = alicloud_nat_gateway.main.forward_table_ids
}
