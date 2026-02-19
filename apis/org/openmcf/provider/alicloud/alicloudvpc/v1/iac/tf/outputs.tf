output "vpc_id" {
  description = "The VPC ID"
  value       = alicloud_vpc.main.id
}

output "vpc_name" {
  description = "The VPC name"
  value       = alicloud_vpc.main.vpc_name
}

output "cidr_block" {
  description = "The primary IPv4 CIDR block of the VPC"
  value       = alicloud_vpc.main.cidr_block
}

output "router_id" {
  description = "The virtual router ID"
  value       = alicloud_vpc.main.router_id
}

output "route_table_id" {
  description = "The system route table ID"
  value       = alicloud_vpc.main.route_table_id
}
