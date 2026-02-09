# outputs.tf

output "monitor_id" {
  description = "The unique identifier (UUID) of the health monitor"
  value       = openstack_lb_monitor_v2.main.id
}

output "name" {
  description = "The name of the health monitor"
  value       = openstack_lb_monitor_v2.main.name
}

output "type" {
  description = "The type of health check"
  value       = openstack_lb_monitor_v2.main.type
}

output "pool_id" {
  description = "The ID of the monitored pool"
  value       = openstack_lb_monitor_v2.main.pool_id
}

output "region" {
  description = "The OpenStack region where the monitor was created"
  value       = openstack_lb_monitor_v2.main.region
}
