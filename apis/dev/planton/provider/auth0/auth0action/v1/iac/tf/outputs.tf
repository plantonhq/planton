# Auth0Action Outputs
# Maps to the Auth0ActionStackOutputs protobuf message

output "id" {
  description = "The unique identifier of the Auth0 action"
  value       = auth0_action.this.id
}

output "name" {
  description = "The name of the action"
  value       = auth0_action.this.name
}

output "version_id" {
  description = "The deployed version ID (when deploy is true)"
  value       = auth0_action.this.version_id
}

output "runtime" {
  description = "The resolved Node.js runtime version"
  value       = auth0_action.this.runtime
}
