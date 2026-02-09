# outputs.tf

output "id" {
  description = "The UUID of the application credential"
  value       = openstack_identity_application_credential_v3.main.id
}

output "name" {
  description = "The name of the application credential"
  value       = openstack_identity_application_credential_v3.main.name
}

output "secret" {
  description = "The application credential secret (SENSITIVE)"
  value       = openstack_identity_application_credential_v3.main.secret
  sensitive   = true
}

output "project_id" {
  description = "The project this credential is scoped to"
  value       = openstack_identity_application_credential_v3.main.project_id
}

output "region" {
  description = "The OpenStack region"
  value       = openstack_identity_application_credential_v3.main.region
}
