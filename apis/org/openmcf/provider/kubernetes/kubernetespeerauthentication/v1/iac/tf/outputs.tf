output "peer_authentication_name" {
  description = "Name of the created PeerAuthentication (equals metadata.name)."
  value       = var.metadata.name
}

output "namespace" {
  description = "Namespace of the created PeerAuthentication."
  value       = var.spec.namespace
}
