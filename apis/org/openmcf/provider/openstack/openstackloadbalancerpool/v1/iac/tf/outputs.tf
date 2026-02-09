# outputs.tf

output "pool_id" {
  description = "The unique identifier (UUID) of the pool"
  value       = openstack_lb_pool_v2.main.id
}

output "name" {
  description = "The name of the pool"
  value       = openstack_lb_pool_v2.main.name
}

output "protocol" {
  description = "The protocol used by pool members"
  value       = openstack_lb_pool_v2.main.protocol
}

output "lb_method" {
  description = "The load-balancing algorithm"
  value       = openstack_lb_pool_v2.main.lb_method
}

output "region" {
  description = "The OpenStack region where the pool was created"
  value       = openstack_lb_pool_v2.main.region
}
