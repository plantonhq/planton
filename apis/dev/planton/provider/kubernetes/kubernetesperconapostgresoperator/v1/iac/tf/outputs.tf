output "namespace" {
  description = "The Kubernetes namespace where the Percona Operator for PostgreSQL is installed"
  value       = helm_release.kubernetes_percona_postgres_operator.namespace
}

output "operator_version" {
  description = "The version of the Percona Operator for PostgreSQL Helm chart deployed"
  value       = local.helm_chart_version
}

output "operator_name" {
  description = "The name of the Percona Operator for PostgreSQL Helm release"
  value       = helm_release.kubernetes_percona_postgres_operator.name
}

output "helm_status" {
  description = "The status of the Helm release deployment"
  value       = helm_release.kubernetes_percona_postgres_operator.status
}

