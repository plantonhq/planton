output "file_system_id" {
  description = "The NAS file system ID"
  value       = alicloud_nas_file_system.main.id
}

output "mount_target_domain" {
  description = "The mount target domain name for NFS/SMB mounting"
  value       = alicloud_nas_mount_target.main.mount_target_domain
}
