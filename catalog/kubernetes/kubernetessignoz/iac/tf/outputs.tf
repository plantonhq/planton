# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports. Every child name derives from the
# fullname pinned via fullnameOverride (the release name). The
# clickhouse_* outputs are passthroughs of the DECLARED connection —
# this component installs no ClickHouse; downstream kinds referencing
# them compose against the same store SigNoz uses.

output "namespace" {
  description = "Kubernetes namespace SigNoz runs in"
  value       = local.namespace
}

output "service" {
  description = "Name of the SigNoz server Service (UI + API, port 8080)"
  value       = local.release_name
}

output "kube_endpoint" {
  description = "In-cluster endpoint of the SigNoz UI/API"
  value       = "http://${local.release_name}.${local.namespace}.svc.cluster.local:8080"
}

output "port_forward_command" {
  description = "kubectl one-liner for opening the SigNoz UI from a workstation (then http://localhost:8080)"
  value       = "kubectl port-forward svc/${local.release_name} -n ${local.namespace} 8080:8080"
}

output "otel_collector_service" {
  description = "Name of the ingestion collector Service (<name>-otel-collector)"
  value       = "${local.release_name}-otel-collector"
}

output "otlp_grpc_endpoint" {
  description = "In-cluster OTLP gRPC ingestion endpoint — point OTLP/gRPC exporters here"
  value       = "${local.release_name}-otel-collector.${local.namespace}.svc.cluster.local:4317"
}

output "otlp_http_endpoint" {
  description = "In-cluster OTLP HTTP ingestion endpoint — point OTLP/HTTP exporters here"
  value       = "http://${local.release_name}-otel-collector.${local.namespace}.svc.cluster.local:4318"
}

output "clickhouse_endpoint" {
  description = "ClickHouse native-protocol endpoint SigNoz stores telemetry in (mirrors the declared connection)"
  value       = "${local.clickhouse.host}:${local.clickhouse_tcp_port}"
}

output "clickhouse_username" {
  description = "ClickHouse username SigNoz connects as (mirrors the declared connection)"
  value       = local.clickhouse.username
}

output "clickhouse_password_secret" {
  description = "Secret key holding the ClickHouse password (the declared Secret reference — the Secret lives in the SigNoz namespace)"
  value = {
    name = local.clickhouse.password_secret.secret_name
    key  = local.clickhouse.password_secret.secret_key
  }
}
