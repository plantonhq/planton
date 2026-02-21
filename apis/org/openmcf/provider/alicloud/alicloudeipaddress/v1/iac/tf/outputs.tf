output "eip_id" {
  description = "The EIP allocation ID"
  value       = alicloud_eip_address.main.id
}

output "ip_address" {
  description = "The allocated public IPv4 address"
  value       = alicloud_eip_address.main.ip_address
}
