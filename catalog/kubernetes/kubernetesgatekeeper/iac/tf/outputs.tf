# Outputs — identical names and derivations in the Pulumi module
# (KubernetesGatekeeperOutputs).

output "namespace" {
  description = "Namespace the engine is installed into"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name (equals metadata.name)"
  value       = local.release_name
}

output "webhook_service_name" {
  description = "Name of the webhook Service the webhook configurations point at (chart-fixed: gatekeeper-webhook-service)"
  value       = local.webhook_service_name
}

output "webhook_cert_secret_name" {
  description = "Name of the Secret carrying the webhook server certificate (chart-fixed: gatekeeper-webhook-server-cert); rotated by the embedded cert-controller unless an external certificate is injected"
  value       = local.webhook_cert_secret_name
}
