# outputs.tf

output "instance_id" {
  description = "The UUID of the instance in OpenStack"
  value       = openstack_compute_instance_v2.main.id
}

output "name" {
  description = "The name of the instance"
  value       = openstack_compute_instance_v2.main.name
}

output "access_ip_v4" {
  description = "The best IPv4 address for accessing the instance"
  value       = openstack_compute_instance_v2.main.access_ip_v4
}

output "access_ip_v6" {
  description = "The best IPv6 address for accessing the instance"
  value       = openstack_compute_instance_v2.main.access_ip_v6
}

output "region" {
  description = "The OpenStack region where the instance was created"
  value       = openstack_compute_instance_v2.main.region
}
