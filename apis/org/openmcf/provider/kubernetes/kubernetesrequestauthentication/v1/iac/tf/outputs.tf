output "request_authentication_name" {
  description = "Name of the created RequestAuthentication (equals metadata.name)"
  value       = local.request_authentication_name
}

output "namespace" {
  description = "Namespace the RequestAuthentication was created in"
  value       = local.namespace
}
