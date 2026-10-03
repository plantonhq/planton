# Semantic outputs mirroring GcpFilestoreInstanceOutputs — names and
# shapes byte-identical to the Pulumi module's exports.

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

# GCP resolves the range even when the spec left it to auto-pick.
output "reserved_ip_range" {
  description = "The /29 CIDR block reserved for this instance"
  value       = try(google_filestore_instance.this.networks[0].reserved_ip_range, "")
}

output "etag" {
  description = "Server-specified ETag guarding concurrent updates"
  value       = google_filestore_instance.this.etag
}
