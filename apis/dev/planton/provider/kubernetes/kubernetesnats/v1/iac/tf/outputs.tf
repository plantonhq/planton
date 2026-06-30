output "namespace" {
  description = "The Kubernetes namespace where NATS is deployed."
  value       = local.namespace
}

output "internal_client_url" {
  description = "The internal NATS client URL for cluster-local connections."
  value       = local.internal_client_url
}

output "external_hostname" {
  description = "The external hostname for NATS if ingress is enabled."
  value       = local.ingress_hostname
}

output "auth_secret_name" {
  description = "The name of the Kubernetes secret containing authentication credentials."
  value       = local.auth_secret_name
}

output "tls_secret_name" {
  description = "The name of the Kubernetes secret containing TLS certificates (if TLS is enabled)."
  value       = try(var.spec.tls_enabled, false) ? local.tls_secret_name : null
}

# NACK Controller outputs
output "nack_controller_enabled" {
  description = "Whether the NACK JetStream controller is enabled."
  value       = local.nack_controller_enabled
}

output "nack_controller_version" {
  description = "The version of the NACK controller Helm chart."
  value       = local.nack_controller_enabled ? local.nack_helm_chart_version : null
}

output "nack_app_version" {
  description = "The NACK app version (GitHub release tag) used for CRDs."
  value       = local.nack_controller_enabled ? local.nack_app_version : null
}

# Streams output
output "streams_created" {
  description = "List of JetStream stream names created by the module."
  value       = local.nack_controller_enabled ? [for stream in local.streams : stream.name] : []
}

# JetStream domain
output "jetstream_domain" {
  description = "The JetStream domain (namespace-based)."
  value       = var.spec.disable_jet_stream ? null : local.namespace
}
