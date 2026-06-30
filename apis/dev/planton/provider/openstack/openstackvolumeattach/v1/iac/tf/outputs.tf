# outputs.tf

output "id" {
  description = "The Terraform resource ID of the volume attachment"
  value       = openstack_compute_volume_attach_v2.main.id
}

output "instance_id" {
  description = "The UUID of the instance the volume is attached to"
  value       = openstack_compute_volume_attach_v2.main.instance_id
}

output "volume_id" {
  description = "The UUID of the Cinder volume that was attached"
  value       = openstack_compute_volume_attach_v2.main.volume_id
}

output "device" {
  description = "The device path where the volume appears in the instance"
  value       = openstack_compute_volume_attach_v2.main.device
}

output "region" {
  description = "The OpenStack region where the attachment was created"
  value       = openstack_compute_volume_attach_v2.main.region
}
