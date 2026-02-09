# outputs.tf

output "network_id" {
  description = "The unique identifier (UUID) of the network"
  value       = openstack_networking_network_v2.main.id
}

output "name" {
  description = "The name of the network"
  value       = openstack_networking_network_v2.main.name
}

output "region" {
  description = "The OpenStack region where the network was created"
  value       = openstack_networking_network_v2.main.region
}
