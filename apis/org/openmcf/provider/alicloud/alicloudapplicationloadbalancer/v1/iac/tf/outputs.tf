output "load_balancer_id" {
  description = "The ALB instance ID"
  value       = alicloud_alb_load_balancer.main.id
}

output "dns_name" {
  description = "The DNS name assigned to the ALB"
  value       = alicloud_alb_load_balancer.main.dns_name
}

output "server_group_ids" {
  description = "Map of server group names to their IDs"
  value = {
    for name, sg in alicloud_alb_server_group.groups : name => sg.id
  }
}
