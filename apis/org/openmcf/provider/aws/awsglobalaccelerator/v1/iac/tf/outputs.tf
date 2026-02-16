output "accelerator_arn" {
  description = "ARN of the Global Accelerator"
  value       = aws_globalaccelerator_accelerator.this.id
}

output "accelerator_dns_name" {
  description = "DNS name of the Global Accelerator"
  value       = aws_globalaccelerator_accelerator.this.dns_name
}

output "accelerator_dual_stack_dns_name" {
  description = "Dual-stack DNS name (IPv4 + IPv6)"
  value       = aws_globalaccelerator_accelerator.this.dual_stack_dns_name
}

output "accelerator_hosted_zone_id" {
  description = "Route53 hosted zone ID for alias records"
  value       = aws_globalaccelerator_accelerator.this.hosted_zone_id
}

output "accelerator_ip_addresses" {
  description = "Static anycast IP addresses assigned to the accelerator"
  value       = flatten([for ip_set in aws_globalaccelerator_accelerator.this.ip_sets : ip_set.ip_addresses])
}

output "listener_arns" {
  description = "Map of listener name to listener ARN"
  value       = { for name, listener in aws_globalaccelerator_listener.this : name => listener.id }
}

output "endpoint_group_arns" {
  description = "Map of composite key (listener_name/group_name) to endpoint group ARN"
  value       = { for key, group in aws_globalaccelerator_endpoint_group.this : key => group.id }
}
