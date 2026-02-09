# outputs.tf

output "server_group_id" {
  description = "The UUID of the server group in OpenStack"
  value       = openstack_compute_servergroup_v2.main.id
}

output "name" {
  description = "The name of the server group"
  value       = openstack_compute_servergroup_v2.main.name
}

output "members" {
  description = "The list of instance UUIDs that belong to this server group"
  value       = openstack_compute_servergroup_v2.main.members
}

output "region" {
  description = "The OpenStack region where the server group was created"
  value       = openstack_compute_servergroup_v2.main.region
}
