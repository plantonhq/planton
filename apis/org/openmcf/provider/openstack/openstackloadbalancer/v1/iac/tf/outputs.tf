# outputs.tf

output "loadbalancer_id" {
  description = "The unique identifier (UUID) of the load balancer"
  value       = openstack_lb_loadbalancer_v2.main.id
}

output "name" {
  description = "The name of the load balancer"
  value       = openstack_lb_loadbalancer_v2.main.name
}

output "vip_address" {
  description = "The Virtual IP address of the load balancer"
  value       = openstack_lb_loadbalancer_v2.main.vip_address
}

output "vip_port_id" {
  description = "The Neutron port ID of the VIP"
  value       = openstack_lb_loadbalancer_v2.main.vip_port_id
}

output "vip_subnet_id" {
  description = "The subnet where the VIP was allocated"
  value       = openstack_lb_loadbalancer_v2.main.vip_subnet_id
}

output "region" {
  description = "The OpenStack region where the load balancer was created"
  value       = openstack_lb_loadbalancer_v2.main.region
}
