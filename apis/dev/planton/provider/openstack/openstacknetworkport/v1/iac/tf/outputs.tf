# outputs.tf

output "port_id" {
  description = "The unique identifier (UUID) of the port resource"
  value       = openstack_networking_port_v2.main.id
}

output "mac_address" {
  description = "The MAC address assigned to the port"
  value       = openstack_networking_port_v2.main.mac_address
}

output "all_fixed_ips" {
  description = "All IP addresses assigned to this port"
  value       = openstack_networking_port_v2.main.all_fixed_ips
}

output "all_security_group_ids" {
  description = "All security group UUIDs applied to this port"
  value       = openstack_networking_port_v2.main.all_security_group_ids
}

output "region" {
  description = "The OpenStack region where the port was created"
  value       = openstack_networking_port_v2.main.region
}
