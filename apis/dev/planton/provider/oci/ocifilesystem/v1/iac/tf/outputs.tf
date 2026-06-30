output "file_system_id" {
  description = "OCID of the file system"
  value       = oci_file_storage_file_system.this.id
}

output "mount_target_id" {
  description = "OCID of the mount target"
  value       = oci_file_storage_mount_target.this.id
}

output "mount_target_ip_address" {
  description = "Private IP address of the mount target for NFS mount commands"
  value       = oci_file_storage_mount_target.this.ip_address
}

output "export_set_id" {
  description = "OCID of the export set associated with the mount target"
  value       = oci_file_storage_mount_target.this.export_set_id
}
