# outputs.tf

output "listener_id" {
  description = "The unique identifier (UUID) of the listener"
  value       = openstack_lb_listener_v2.main.id
}

output "name" {
  description = "The name of the listener"
  value       = openstack_lb_listener_v2.main.name
}

output "protocol" {
  description = "The protocol the listener accepts"
  value       = openstack_lb_listener_v2.main.protocol
}

output "protocol_port" {
  description = "The port on which the listener accepts traffic"
  value       = openstack_lb_listener_v2.main.protocol_port
}

output "region" {
  description = "The OpenStack region where the listener was created"
  value       = openstack_lb_listener_v2.main.region
}
