# outputs.tf

output "id" {
  description = "The composite identifier of the role assignment"
  value       = openstack_identity_role_assignment_v3.main.id
}

output "role_id" {
  description = "The UUID of the assigned role"
  value       = openstack_identity_role_assignment_v3.main.role_id
}

output "project_id" {
  description = "The project scope (if project-scoped)"
  value       = openstack_identity_role_assignment_v3.main.project_id
}

output "domain_id" {
  description = "The domain scope (if domain-scoped)"
  value       = openstack_identity_role_assignment_v3.main.domain_id
}

output "user_id" {
  description = "The user principal (if user assignment)"
  value       = openstack_identity_role_assignment_v3.main.user_id
}

output "group_id" {
  description = "The group principal (if group assignment)"
  value       = openstack_identity_role_assignment_v3.main.group_id
}

output "region" {
  description = "The OpenStack region"
  value       = openstack_identity_role_assignment_v3.main.region
}
