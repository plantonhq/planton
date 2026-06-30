output "namespace" {
  description = "Namespace where the Certificate resource was created"
  value       = local.namespace
}

output "certificate_name" {
  description = "Name of the created Certificate resource"
  value       = local.certificate_name
}

output "secret_name" {
  description = "TLS Secret name containing the signed certificate and private key"
  value       = local.secret_name
}
