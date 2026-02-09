# outputs.tf

output "id" {
  description = "The Terraform resource ID of the association"
  value       = openstack_networking_floatingip_associate_v2.main.id
}

output "floating_ip" {
  description = "The floating IP address that was associated"
  value       = openstack_networking_floatingip_associate_v2.main.floating_ip
}

output "port_id" {
  description = "The UUID of the port the floating IP was associated with"
  value       = openstack_networking_floatingip_associate_v2.main.port_id
}

output "fixed_ip" {
  description = "The fixed IP on the port that the floating IP maps to"
  value       = openstack_networking_floatingip_associate_v2.main.fixed_ip
}

output "region" {
  description = "The OpenStack region where the association was created"
  value       = openstack_networking_floatingip_associate_v2.main.region
}
