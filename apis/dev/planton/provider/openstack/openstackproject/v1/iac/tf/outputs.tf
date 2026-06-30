# outputs.tf

output "project_id" {
  description = "The UUID of the project in OpenStack"
  value       = openstack_identity_project_v3.main.id
}

output "name" {
  description = "The name of the project"
  value       = openstack_identity_project_v3.main.name
}

output "domain_id" {
  description = "The Keystone domain the project belongs to"
  value       = openstack_identity_project_v3.main.domain_id
}

output "enabled" {
  description = "Whether the project is active"
  value       = openstack_identity_project_v3.main.enabled
}

output "region" {
  description = "The OpenStack region where the project was created"
  value       = openstack_identity_project_v3.main.region
}
