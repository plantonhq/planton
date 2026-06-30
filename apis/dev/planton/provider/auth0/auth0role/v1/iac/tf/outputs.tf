# Auth0Role Outputs
# Maps to the Auth0RoleStackOutputs protobuf message

output "id" {
  description = "The unique identifier of the Auth0 role (e.g. rol_abc123)"
  value       = auth0_role.this.id
}

output "name" {
  description = "The human-readable name of the role"
  value       = auth0_role.this.name
}

output "description" {
  description = "The human-readable description of the role"
  value       = auth0_role.this.description
}
