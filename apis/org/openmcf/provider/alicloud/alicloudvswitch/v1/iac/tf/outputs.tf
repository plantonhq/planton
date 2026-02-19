output "vswitch_id" {
  description = "The VSwitch ID"
  value       = alicloud_vswitch.main.id
}

output "vswitch_name" {
  description = "The VSwitch name"
  value       = alicloud_vswitch.main.vswitch_name
}

output "cidr_block" {
  description = "The IPv4 CIDR block of the VSwitch"
  value       = alicloud_vswitch.main.cidr_block
}

output "zone_id" {
  description = "The availability zone"
  value       = alicloud_vswitch.main.zone_id
}

output "ipv6_cidr_block" {
  description = "The IPv6 CIDR block (empty if IPv6 is not enabled)"
  value       = alicloud_vswitch.main.ipv6_cidr_block
}
