# outputs.tf

output "security_group_id" {
  description = "The unique identifier (UUID) of the security group"
  value       = openstack_networking_secgroup_v2.main.id
}

output "name" {
  description = "The name of the security group"
  value       = openstack_networking_secgroup_v2.main.name
}

output "region" {
  description = "The OpenStack region where the security group was created"
  value       = openstack_networking_secgroup_v2.main.region
}
