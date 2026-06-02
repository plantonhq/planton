output "authorization_policy_name" {
  description = "Name of the created AuthorizationPolicy (equals metadata.name)"
  value       = local.authorization_policy_name
}

output "namespace" {
  description = "Namespace the AuthorizationPolicy was created in"
  value       = local.namespace
}
