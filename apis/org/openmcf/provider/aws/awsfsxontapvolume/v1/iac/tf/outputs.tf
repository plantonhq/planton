# ---------------------------------------------------------------------------
# Volume Identification
# ---------------------------------------------------------------------------

output "volume_id" {
  description = "The ID of the ONTAP volume (e.g., fsvol-0123456789abcdef0)."
  value       = aws_fsx_ontap_volume.this.id
}

output "arn" {
  description = "The Amazon Resource Name of the volume."
  value       = aws_fsx_ontap_volume.this.arn
}

output "uuid" {
  description = "The volume's UUID in ONTAP."
  value       = aws_fsx_ontap_volume.this.uuid
}

output "file_system_id" {
  description = "The file system ID that this volume belongs to."
  value       = aws_fsx_ontap_volume.this.file_system_id
}

output "flexcache_endpoint_type" {
  description = "The FlexCache endpoint type (NONE, ORIGIN, or CACHE)."
  value       = aws_fsx_ontap_volume.this.flexcache_endpoint_type
}

output "ontap_volume_type" {
  description = "The ONTAP volume type (RW or DP)."
  value       = aws_fsx_ontap_volume.this.ontap_volume_type
}
