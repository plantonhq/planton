output "namespace" {
  description = "The namespace in which SigNoz is deployed."
  value       = local.namespace
}

output "signoz_service" {
  description = "The name of the SigNoz UI and API service."
  value       = local.signoz_service_name
}

output "otel_collector_service" {
  description = "The name of the OpenTelemetry Collector service."
  value       = local.otel_collector_service_name
}

output "kube_endpoint" {
  description = "Internal DNS name of the SigNoz UI service within the cluster."
  value       = local.signoz_kube_endpoint
}

output "otel_collector_grpc_endpoint" {
  description = "Internal DNS name of the OpenTelemetry Collector gRPC endpoint."
  value       = local.otel_collector_grpc_endpoint
}

output "otel_collector_http_endpoint" {
  description = "Internal DNS name of the OpenTelemetry Collector HTTP endpoint."
  value       = local.otel_collector_http_endpoint
}

output "port_forward_command" {
  description = "Handy command to port-forward traffic to the SigNoz UI on localhost:8080."
  value       = "kubectl port-forward -n ${local.namespace} service/${local.signoz_service_name} ${local.signoz_ui_port}:${local.signoz_ui_port}"
}

output "external_hostname" {
  description = "The external hostname for SigNoz UI if ingress is enabled."
  value       = local.signoz_ingress_external_hostname
}

output "otel_collector_external_http_hostname" {
  description = "The external HTTP hostname for OpenTelemetry Collector if ingress is enabled."
  value       = local.otel_collector_external_http_hostname
}

output "clickhouse_endpoint" {
  description = "The internal ClickHouse endpoint (only for self-managed mode)."
  value       = local.clickhouse_endpoint
}

output "clickhouse_username" {
  description = "The ClickHouse username (only for self-managed mode)."
  value       = !local.is_external_database ? "admin" : null
}

output "clickhouse_password_secret_name" {
  description = "Name of the Secret holding the ClickHouse password (only for self-managed mode)."
  value       = !local.is_external_database ? "${var.metadata.name}-clickhouse" : null
}

output "clickhouse_password_secret_key" {
  description = "Key within the Secret that contains the ClickHouse password (only for self-managed mode)."
  value       = !local.is_external_database ? "admin-password" : null
}

