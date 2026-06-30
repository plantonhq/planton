output "aws_ecs_service_name" {
  description = "The final name of the ECS service"
  value       = aws_ecs_service.this.name
}

output "ecs_cluster_name" {
  description = "Indicates which cluster the service is deployed in"
  value       = local.cluster_name
}

output "load_balancer_dns_name" {
  description = "The DNS name of the ALB/NLB if ALB is configured"
  value       = local.has_alb_config ? data.aws_lb.selected[0].dns_name : null
}

output "service_url" {
  description = "The final external endpoint if ALB is configured"
  value       = local.has_alb_config ? "http://${data.aws_lb.selected[0].dns_name}" : null
}

output "service_discovery_name" {
  description = "The internal DNS name if service discovery was used"
  value       = null
}

output "cloudwatch_log_group_name" {
  description = "The name of the CloudWatch log group for the service"
  value       = local.logging_enabled ? aws_cloudwatch_log_group.this[0].name : null
}

output "cloudwatch_log_group_arn" {
  description = "The ARN of the CloudWatch log group for the service"
  value       = local.logging_enabled ? aws_cloudwatch_log_group.this[0].arn : null
}

output "service_arn" {
  description = "The Amazon Resource Name of the ECS service"
  value       = aws_ecs_service.this.id
}

output "task_definition_arn" {
  description = "The ARN of the ECS task definition"
  value       = aws_ecs_task_definition.this.arn
}

output "target_group_arn" {
  description = "The ARN of the associated target group when ALB is enabled"
  value       = local.has_alb_config && local.has_container_port ? aws_lb_target_group.this[0].arn : null
}

output "target_group_name" {
  description = "The name of the associated target group when ALB is enabled"
  value       = local.has_alb_config && local.has_container_port ? aws_lb_target_group.this[0].name : null
}


