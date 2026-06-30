output "cluster_identifier" {
  description = "Unique identifier of the Redshift cluster"
  value       = aws_redshift_cluster.this.cluster_identifier
}

output "cluster_arn" {
  description = "ARN of the Redshift cluster"
  value       = aws_redshift_cluster.this.arn
}

output "cluster_namespace_arn" {
  description = "Namespace ARN for data sharing and Redshift Serverless integration"
  value       = aws_redshift_cluster.this.cluster_namespace_arn
}

output "endpoint" {
  description = "Connection endpoint in address:port format"
  value       = aws_redshift_cluster.this.endpoint
}

output "dns_name" {
  description = "DNS hostname of the cluster (without port)"
  value       = aws_redshift_cluster.this.dns_name
}

output "database_name" {
  description = "Name of the default database"
  value       = aws_redshift_cluster.this.database_name
}

output "port" {
  description = "TCP port for client connections"
  value       = aws_redshift_cluster.this.port
}

output "subnet_group_name" {
  description = "Name of the managed Redshift subnet group (empty if not created)"
  value       = local.create_subnet_group ? aws_redshift_subnet_group.this[0].name : ""
}

output "security_group_id" {
  description = "ID of the managed security group (empty if not created)"
  value       = local.create_security_group ? aws_security_group.this[0].id : ""
}

output "parameter_group_name" {
  description = "Name of the managed parameter group (empty if not created)"
  value       = local.create_parameter_group ? aws_redshift_parameter_group.this[0].name : ""
}

output "master_password_secret_arn" {
  description = "ARN of the Secrets Manager secret containing the master password (empty when manage_master_password is false)"
  value       = try(aws_redshift_cluster.this.master_password_secret_arn, "")
}
