output "namespace" {
  description = "The Kubernetes namespace where the Percona Operator for MySQL is installed"
  value       = local.namespace
}

output "operator_version" {
  description = "The version of the Percona Operator for MySQL Helm chart deployed"
  value       = local.helm_chart_version
}

output "operator_name" {
  description = "The name of the Percona Operator for MySQL Helm release"
  value       = helm_release.percona_mysql_operator.name
}

output "helm_status" {
  description = "The status of the Helm release deployment"
  value       = helm_release.percona_mysql_operator.status
}

