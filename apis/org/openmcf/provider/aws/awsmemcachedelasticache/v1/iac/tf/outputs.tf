output "cluster_id" {
  description = "The identifier of the ElastiCache cluster."
  value       = aws_elasticache_cluster.this.cluster_id
}

output "cluster_address" {
  description = "The DNS name of the Memcached auto-discovery endpoint (without port)."
  value       = aws_elasticache_cluster.this.cluster_address
}

output "configuration_endpoint" {
  description = "The full configuration endpoint (address:port) for client auto-discovery."
  value       = aws_elasticache_cluster.this.configuration_endpoint
}

output "arn" {
  description = "The ARN of the ElastiCache cluster."
  value       = aws_elasticache_cluster.this.arn
}

output "port" {
  description = "The port on which the cluster accepts connections."
  value       = aws_elasticache_cluster.this.port
}

output "subnet_group_name" {
  description = "The name of the created subnet group (empty if none created)."
  value       = local.has_subnets ? aws_elasticache_subnet_group.this[0].name : ""
}

output "parameter_group_name" {
  description = "The name of the created parameter group (empty if none created)."
  value       = local.has_parameters ? aws_elasticache_parameter_group.this[0].name : ""
}
