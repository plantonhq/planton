# Outputs — flattened onto KubernetesOtelCollectorOutputs by
# the platform. Keep in lockstep with the Pulumi module's exports.
#
# Sidecar mode creates NO standalone workload — every Service-derived
# handle is empty there (the operator injects the collector into
# annotated pods instead).

output "namespace" {
  description = "Kubernetes namespace the collector runs in"
  value       = local.namespace
}

output "collector_name" {
  description = "Name of the OpenTelemetryCollector custom resource (metadata.name; the operator derives every child name from it)"
  value       = local.resource_name
}

output "service" {
  description = "Name of the collector Service (\"<name>-collector\") carrying the receiver-derived ports; empty in sidecar mode"
  value       = local.effective_mode == "sidecar" ? "" : "${local.resource_name}-collector"
}

output "otlp_grpc_endpoint" {
  description = "In-cluster OTLP gRPC ingest endpoint (\"<service>:4317\") — valid when the config declares the standard otlp receiver; empty in sidecar mode"
  value       = local.effective_mode == "sidecar" ? "" : "${local.resource_name}-collector.${local.namespace}.svc.cluster.local:4317"
}

output "otlp_http_endpoint" {
  description = "In-cluster OTLP HTTP ingest endpoint (\"http://<service>:4318\") — valid when the config declares the standard otlp receiver; empty in sidecar mode"
  value       = local.effective_mode == "sidecar" ? "" : "http://${local.resource_name}-collector.${local.namespace}.svc.cluster.local:4318"
}

output "headless_service" {
  description = "Name of the headless Service (\"<name>-collector-headless\") for per-pod addressing; empty in sidecar mode"
  value       = local.effective_mode == "sidecar" ? "" : "${local.resource_name}-collector-headless"
}

output "monitoring_service" {
  description = "Name of the monitoring Service (\"<name>-collector-monitoring\", port 8888) — the collector's own metrics; empty in sidecar mode"
  value       = local.effective_mode == "sidecar" ? "" : "${local.resource_name}-collector-monitoring"
}
