# ---------------------------------------------------------------------------
# Outputs — matching AwsFsxOntapFileSystemOutputs
# ---------------------------------------------------------------------------
# Primary consumers: AwsFsxOntapStorageVirtualMachine (file_system_id),
# AwsFsxOntapVolume (via SVM), IAM policies (file_system_arn), SnapMirror
# (intercluster_*), ONTAP CLI (management_*). Data access endpoints (NFS/SMB/
# iSCSI) live on the SVM, not here — an ONTAP file system has no
# file-system-level data DNS name.
# ---------------------------------------------------------------------------

output "file_system_id" {
  description = "The ID of the file system (e.g., fs-0123456789abcdef0)."
  value       = aws_fsx_ontap_file_system.this.id
}

output "file_system_arn" {
  description = "The ARN of the file system for IAM resource-level permissions."
  value       = aws_fsx_ontap_file_system.this.arn
}

output "management_dns_name" {
  description = "Management endpoint DNS name for ONTAP CLI and REST API (ssh fsxadmin@<management_dns_name>)."
  value       = try(aws_fsx_ontap_file_system.this.endpoints[0].management[0].dns_name, "")
}

output "management_ip_addresses" {
  description = "Management endpoint IP addresses for direct ONTAP management access."
  value       = try(aws_fsx_ontap_file_system.this.endpoints[0].management[0].ip_addresses, [])
}

output "intercluster_dns_name" {
  description = "Intercluster endpoint DNS name for NetApp SnapMirror replication."
  value       = try(aws_fsx_ontap_file_system.this.endpoints[0].intercluster[0].dns_name, "")
}

output "intercluster_ip_addresses" {
  description = "Intercluster endpoint IP addresses for SnapMirror peering."
  value       = try(aws_fsx_ontap_file_system.this.endpoints[0].intercluster[0].ip_addresses, [])
}

output "network_interface_ids" {
  description = "Network interface IDs created for the file system."
  value       = aws_fsx_ontap_file_system.this.network_interface_ids
}

output "vpc_id" {
  description = "VPC ID in which the file system was created."
  value       = aws_fsx_ontap_file_system.this.vpc_id
}

output "owner_id" {
  description = "AWS account ID of the file system owner."
  value       = aws_fsx_ontap_file_system.this.owner_id
}
