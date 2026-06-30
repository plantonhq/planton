output "cluster_name" {
  description = "The ECS cluster name."
  value       = aws_ecs_cluster.this.name
}

output "cluster_arn" {
  description = "The ECS cluster ARN."
  value       = aws_ecs_cluster.this.arn
}

output "cluster_capacity_providers" {
  description = "The capacity providers associated with this cluster."
  value       = local.has_capacity_providers ? var.spec.capacity_providers : []
}


