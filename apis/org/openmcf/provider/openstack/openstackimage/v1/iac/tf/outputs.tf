# outputs.tf

output "image_id" {
  description = "The UUID of the image in Glance"
  value       = openstack_images_image_v2.main.id
}

output "name" {
  description = "The name of the image"
  value       = openstack_images_image_v2.main.name
}

output "checksum" {
  description = "The MD5 checksum of the image data"
  value       = openstack_images_image_v2.main.checksum
}

output "size_bytes" {
  description = "The size of the image data in bytes"
  value       = openstack_images_image_v2.main.size_bytes
}

output "status" {
  description = "The lifecycle status of the image"
  value       = openstack_images_image_v2.main.status
}

output "file" {
  description = "The URL path to the image data in Glance"
  value       = openstack_images_image_v2.main.file
}

output "region" {
  description = "The OpenStack region where the image was created"
  value       = openstack_images_image_v2.main.region
}
