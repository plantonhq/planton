# outputs.tf

output "subnet_id" {
  description = "The unique identifier (UUID) of the subnet"
  value       = openstack_networking_subnet_v2.main.id
}

output "name" {
  description = "The name of the subnet"
  value       = openstack_networking_subnet_v2.main.name
}

output "cidr" {
  description = "The CIDR block of the subnet"
  value       = openstack_networking_subnet_v2.main.cidr
}

output "gateway_ip" {
  description = "The gateway IP address of the subnet"
  value       = openstack_networking_subnet_v2.main.gateway_ip
}

output "network_id" {
  description = "The ID of the parent network"
  value       = openstack_networking_subnet_v2.main.network_id
}

output "region" {
  description = "The OpenStack region where the subnet was created"
  value       = openstack_networking_subnet_v2.main.region
}
