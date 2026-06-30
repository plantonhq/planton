output "cluster_arn" {
  description = "ARN of the MSK cluster"
  value       = aws_msk_cluster.this.arn
}

output "cluster_name" {
  description = "Name of the MSK cluster"
  value       = aws_msk_cluster.this.cluster_name
}

output "cluster_uuid" {
  description = "UUID of the MSK cluster (extracted from ARN)"
  value       = aws_msk_cluster.this.cluster_uuid
}

output "current_version" {
  description = "Current version of the MSK cluster"
  value       = aws_msk_cluster.this.current_version
}

output "bootstrap_brokers" {
  description = "Comma-separated list of plaintext bootstrap broker endpoints"
  value       = aws_msk_cluster.this.bootstrap_brokers
}

output "bootstrap_brokers_tls" {
  description = "Comma-separated list of TLS bootstrap broker endpoints"
  value       = aws_msk_cluster.this.bootstrap_brokers_tls
}

output "bootstrap_brokers_sasl_iam" {
  description = "Comma-separated list of SASL/IAM bootstrap broker endpoints"
  value       = aws_msk_cluster.this.bootstrap_brokers_sasl_iam
}

output "bootstrap_brokers_sasl_scram" {
  description = "Comma-separated list of SASL/SCRAM bootstrap broker endpoints"
  value       = aws_msk_cluster.this.bootstrap_brokers_sasl_scram
}

output "bootstrap_brokers_public_tls" {
  description = "Comma-separated list of public TLS bootstrap broker endpoints"
  value       = aws_msk_cluster.this.bootstrap_brokers_public_tls
}

output "bootstrap_brokers_public_sasl_iam" {
  description = "Comma-separated list of public SASL/IAM bootstrap broker endpoints"
  value       = aws_msk_cluster.this.bootstrap_brokers_public_sasl_iam
}

output "bootstrap_brokers_public_sasl_scram" {
  description = "Comma-separated list of public SASL/SCRAM bootstrap broker endpoints"
  value       = aws_msk_cluster.this.bootstrap_brokers_public_sasl_scram
}

output "zookeeper_connect_string" {
  description = "Comma-separated list of ZooKeeper plaintext endpoints"
  value       = aws_msk_cluster.this.zookeeper_connect_string
}

output "zookeeper_connect_string_tls" {
  description = "Comma-separated list of ZooKeeper TLS endpoints"
  value       = aws_msk_cluster.this.zookeeper_connect_string_tls
}

output "security_group_id" {
  description = "ID of the managed security group (if created)"
  value       = length(aws_security_group.cluster) > 0 ? aws_security_group.cluster[0].id : ""
}

output "configuration_arn" {
  description = "ARN of the inline MSK Configuration (if created)"
  value       = length(aws_msk_configuration.inline) > 0 ? aws_msk_configuration.inline[0].arn : ""
}
