# outputs.tf

output "floating_ip_id" {
  description = "The unique identifier (UUID) of the floating IP resource"
  value       = openstack_networking_floatingip_v2.main.id
}

output "address" {
  description = "The allocated floating IP address (e.g., 203.0.113.42)"
  value       = openstack_networking_floatingip_v2.main.address
}

output "floating_network_id" {
  description = "The UUID of the external network the floating IP was allocated from"
  value       = openstack_networking_floatingip_v2.main.pool
}

output "port_id" {
  description = "The UUID of the associated port (empty if allocation-only)"
  value       = openstack_networking_floatingip_v2.main.port_id
}

output "fixed_ip" {
  description = "The fixed IP mapped to this floating IP (empty if no association)"
  value       = openstack_networking_floatingip_v2.main.fixed_ip
}

output "region" {
  description = "The OpenStack region where the floating IP was allocated"
  value       = openstack_networking_floatingip_v2.main.region
}
