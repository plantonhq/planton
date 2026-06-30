# outputs.tf

output "volume_id" {
  description = "The UUID of the Cinder volume"
  value       = openstack_blockstorage_volume_v3.main.id
}

output "name" {
  description = "The name of the volume"
  value       = openstack_blockstorage_volume_v3.main.name
}

output "size" {
  description = "The size of the volume in gigabytes"
  value       = openstack_blockstorage_volume_v3.main.size
}

output "volume_type" {
  description = "The volume type (Cinder backend storage class)"
  value       = openstack_blockstorage_volume_v3.main.volume_type
}

output "availability_zone" {
  description = "The availability zone where the volume was created"
  value       = openstack_blockstorage_volume_v3.main.availability_zone
}

output "region" {
  description = "The OpenStack region where the volume was created"
  value       = openstack_blockstorage_volume_v3.main.region
}
