# Terraform outputs for Kubernetes Secret

output "secret_name" {
  description = "The name of the created Kubernetes Secret"
  value       = kubernetes_secret_v1.secret.metadata[0].name
}

output "secret_namespace" {
  description = "The namespace where the secret was created"
  value       = kubernetes_secret_v1.secret.metadata[0].namespace
}

output "secret_type" {
  description = "The Kubernetes secret type string"
  value       = local.secret_type
}
