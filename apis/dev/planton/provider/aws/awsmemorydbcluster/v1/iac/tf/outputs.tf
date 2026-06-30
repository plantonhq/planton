output "cluster_endpoint_address" {
  description = "The DNS address of the MemoryDB cluster endpoint"
  value       = try(aws_memorydb_cluster.this.cluster_endpoint[0].address, "")
}

output "cluster_endpoint_port" {
  description = "The port of the MemoryDB cluster endpoint"
  value       = try(aws_memorydb_cluster.this.cluster_endpoint[0].port, 6379)
}

output "cluster_arn" {
  description = "The ARN of the MemoryDB cluster"
  value       = aws_memorydb_cluster.this.arn
}

output "cluster_name" {
  description = "The name of the MemoryDB cluster"
  value       = aws_memorydb_cluster.this.name
}

output "engine_patch_version" {
  description = "The actual engine patch version running on the cluster"
  value       = aws_memorydb_cluster.this.engine_patch_version
}

output "subnet_group_name" {
  description = "The name of the created subnet group (empty if not created)"
  value       = local.create_subnet_group ? aws_memorydb_subnet_group.this[0].name : ""
}

output "parameter_group_name" {
  description = "The name of the created parameter group (empty if not created)"
  value       = local.create_parameter_group ? aws_memorydb_parameter_group.this[0].name : ""
}
