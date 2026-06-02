output "peer_authentication_name" {
  description = "Name of the created PeerAuthentication (equals metadata.name)"
  value       = local.peer_authentication_name
}

output "namespace" {
  description = "Namespace the PeerAuthentication was created in"
  value       = local.namespace
}
