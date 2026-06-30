# Outputs matching AwsDocumentDbStackOutputs

output "cluster_endpoint" {
  description = "The primary endpoint for the DocumentDB cluster (writer endpoint)"
  value       = aws_docdb_cluster.main.endpoint
}

output "cluster_reader_endpoint" {
  description = "The reader endpoint for load-balanced read traffic"
  value       = aws_docdb_cluster.main.reader_endpoint
}

output "cluster_id" {
  description = "The AWS identifier of the DocumentDB cluster"
  value       = aws_docdb_cluster.main.id
}

output "cluster_arn" {
  description = "The Amazon Resource Name (ARN) of the DocumentDB cluster"
  value       = aws_docdb_cluster.main.arn
}

output "cluster_port" {
  description = "The port on which the DocumentDB cluster accepts connections"
  value       = aws_docdb_cluster.main.port
}

output "db_subnet_group_name" {
  description = "The name of the DB subnet group"
  value       = local.need_subnet_group ? aws_docdb_subnet_group.main[0].name : local.subnet_group_name_var
}

output "security_group_id" {
  description = "The security group ID associated with the cluster"
  value       = local.need_managed_sg ? aws_security_group.main[0].id : null
}

output "cluster_parameter_group_name" {
  description = "The cluster parameter group name in use"
  value       = local.need_cluster_parameter_group ? aws_docdb_cluster_parameter_group.main[0].name : try(var.spec.cluster_parameter_group_name, null)
}

output "connection_string" {
  description = "MongoDB-compatible connection string template"
  value       = "mongodb://${coalesce(try(var.spec.master_username, ""), "docdbadmin")}:<password>@${aws_docdb_cluster.main.endpoint}:${aws_docdb_cluster.main.port}/?tls=true&replicaSet=rs0&readPreference=secondaryPreferred&retryWrites=false"
}

output "cluster_resource_id" {
  description = "The cluster resource ID (internal AWS identifier)"
  value       = aws_docdb_cluster.main.cluster_resource_id
}
