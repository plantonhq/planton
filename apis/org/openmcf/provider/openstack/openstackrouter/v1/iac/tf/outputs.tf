# outputs.tf

output "router_id" {
  description = "The unique identifier (UUID) of the router"
  value       = openstack_networking_router_v2.main.id
}

output "name" {
  description = "The name of the router"
  value       = openstack_networking_router_v2.main.name
}

output "external_network_id" {
  description = "The ID of the external network (empty if no gateway configured)"
  value       = openstack_networking_router_v2.main.external_network_id
}

output "external_gateway_ip" {
  description = "The primary external IP address of the router's gateway"
  value = (
    length(openstack_networking_router_v2.main.external_fixed_ip) > 0
    ? openstack_networking_router_v2.main.external_fixed_ip[0].ip_address
    : ""
  )
}

output "region" {
  description = "The OpenStack region where the router was created"
  value       = openstack_networking_router_v2.main.region
}
