output "replication_group_id" {
  description = "The identifier of the replication group."
  value       = aws_elasticache_replication_group.this.id
}

output "primary_endpoint_address" {
  description = "The primary (writer) endpoint DNS name."
  value       = aws_elasticache_replication_group.this.primary_endpoint_address
}

output "reader_endpoint_address" {
  description = "The reader endpoint DNS name for read replicas."
  value       = aws_elasticache_replication_group.this.reader_endpoint_address
}

output "configuration_endpoint_address" {
  description = "The configuration endpoint for Cluster Mode Enabled."
  value       = aws_elasticache_replication_group.this.configuration_endpoint_address
}

output "arn" {
  description = "The ARN of the replication group."
  value       = aws_elasticache_replication_group.this.arn
}

output "port" {
  description = "The port on which the cluster accepts connections."
  value       = aws_elasticache_replication_group.this.port
}

output "subnet_group_name" {
  description = "The name of the created subnet group (empty if none created)."
  value       = local.has_subnets ? aws_elasticache_subnet_group.this[0].name : ""
}

output "parameter_group_name" {
  description = "The name of the created parameter group (empty if none created)."
  value       = local.has_parameters ? aws_elasticache_parameter_group.this[0].name : ""
}
