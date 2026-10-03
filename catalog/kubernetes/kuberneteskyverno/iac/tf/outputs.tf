# Outputs — identical names and derivations in the Pulumi module
# (KubernetesKyvernoOutputs).

output "namespace" {
  description = "Namespace the engine is installed into"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name (equals metadata.name)"
  value       = local.release_name
}

output "admission_service_name" {
  description = "Name of the admission controller's webhook Service — the backend the runtime-registered webhook configurations point at (<fullname>-svc)"
  value       = local.admission_service_name
}

output "config_map_name" {
  description = "Name of the engine's runtime ConfigMap (resource filters, webhook selectors) — the object to inspect when a resource is unexpectedly skipped or policed"
  value       = local.config_map_name
}
