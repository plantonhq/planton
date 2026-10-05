# Outputs — flattened onto KubernetesOtelOperatorOutputs by the
# platform. Keep in lockstep with the Pulumi module's exports.

output "namespace" {
  description = "Kubernetes namespace the operator was installed into"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name of the operator (metadata.name; the chart's fullname is pinned to it)"
  value       = local.release_name
}

output "webhook_service" {
  description = "Name of the operator's webhook Service (\"<name>-webhook\", port 443) — where the API server sends admission reviews and CRD conversion calls"
  value       = local.webhook_service
}

output "webhook_cert_secret_name" {
  description = "Name of the Secret holding the webhook serving certificate (\"<name>-controller-manager-service-cert\", cert-manager-issued)"
  value       = local.webhook_cert_secret_name
}
