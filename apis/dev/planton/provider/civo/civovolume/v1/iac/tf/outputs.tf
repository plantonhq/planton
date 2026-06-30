output "volume_id" {
  description = "The unique identifier (ID) of the created Civo volume"
  value       = civo_volume.main.id
}

output "attached_instance_id" {
  description = "The ID of the Civo instance the volume is attached to (if any). Empty if unattached."
  value       = try(civo_volume.main.attached_instance_id, "")
}

output "device_path" {
  description = "The device path of the volume on the attached instance (if any). Example: /dev/vdb"
  value       = ""
  # Note: The Civo provider doesn't expose the device path.
  # Users need to identify the device after attachment, typically by checking
  # /dev/disk/by-id/ or using lsblk on the instance.
}

output "volume_name" {
  description = "The name of the created volume"
  value       = civo_volume.main.name
}

output "size_gib" {
  description = "The size of the volume in GiB"
  value       = civo_volume.main.size_gb
}

output "region" {
  description = "The region where the volume was created"
  value       = local.region
}

output "mount_point" {
  description = "The mount point on the instance (informational only - set by user after attachment)"
  value       = ""
  # Note: Mount point is configured by the user after attaching and formatting the volume.
  # Common mount points: /data, /mnt/data, /var/lib/mysql, etc.
}

