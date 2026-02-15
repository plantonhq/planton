output "load_balancer_arn" {
  description = "ARN of the Network Load Balancer"
  value       = aws_lb.this.arn
}

output "load_balancer_name" {
  description = "Name of the Network Load Balancer"
  value       = aws_lb.this.name
}

output "load_balancer_dns_name" {
  description = "DNS name assigned by AWS to the Network Load Balancer"
  value       = aws_lb.this.dns_name
}

output "load_balancer_hosted_zone_id" {
  description = "Route53 hosted zone ID for the NLB's DNS name"
  value       = aws_lb.this.zone_id
}

output "listener_arns" {
  description = "Map of listener name to listener ARN"
  value       = { for k, v in aws_lb_listener.this : k => v.arn }
}

output "target_group_arns" {
  description = "Map of listener name to target group ARN"
  value       = { for k, v in aws_lb_target_group.this : k => v.arn }
}
