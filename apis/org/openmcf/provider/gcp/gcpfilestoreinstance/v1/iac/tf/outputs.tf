output "instance_id" {
  description = "Fully qualified resource ID of the Filestore instance"
  value       = google_filestore_instance.this.id
}

output "instance_name" {
  description = "Short name of the Filestore instance"
  value       = google_filestore_instance.this.name
}

output "ip_addresses" {
  description = "IP addresses assigned to the instance on its VPC network"
  value       = try(google_filestore_instance.this.networks[0].ip_addresses, [])
}

output "file_share_name" {
  description = "Name of the file share (for NFS mount path)"
  value       = var.spec.file_share.name
}

output "create_time" {
  description = "Timestamp when the instance was created (RFC3339 format)"
  value       = google_filestore_instance.this.create_time
}
