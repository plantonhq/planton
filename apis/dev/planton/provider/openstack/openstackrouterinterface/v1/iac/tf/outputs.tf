# outputs.tf

output "port_id" {
  description = "The UUID of the port created by the router interface attachment"
  value       = openstack_networking_router_interface_v2.main.id
}

output "router_id" {
  description = "The UUID of the router this interface is attached to"
  value       = openstack_networking_router_interface_v2.main.router_id
}

output "subnet_id" {
  description = "The UUID of the subnet connected to the router"
  value       = openstack_networking_router_interface_v2.main.subnet_id
}

output "region" {
  description = "The OpenStack region where the router interface was created"
  value       = openstack_networking_router_interface_v2.main.region
}
