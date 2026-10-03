# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports.

output "namespace" {
  description = "Namespace the Certificate resource was created in"
  value       = local.namespace
}

output "certificate_name" {
  description = "Name of the created Certificate resource"
  value       = local.certificate_name
}

output "secret_name" {
  description = "TLS Secret name — the handle consumers reference (Ingress tls.secretName, Gateway certificateRefs, CA Issuer ca_secret_name)"
  value       = local.secret_name
}
