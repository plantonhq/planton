# ---------------------------------------------------------------------------
# SVM Identification
# ---------------------------------------------------------------------------

output "svm_id" {
  description = "The ID of the Storage Virtual Machine."
  value       = aws_fsx_ontap_storage_virtual_machine.this.id
}

output "arn" {
  description = "The Amazon Resource Name of the SVM."
  value       = aws_fsx_ontap_storage_virtual_machine.this.arn
}

output "uuid" {
  description = "The SVM's UUID in ONTAP."
  value       = aws_fsx_ontap_storage_virtual_machine.this.uuid
}

output "subtype" {
  description = "The SVM subtype (e.g., DEFAULT)."
  value       = aws_fsx_ontap_storage_virtual_machine.this.subtype
}

# ---------------------------------------------------------------------------
# iSCSI Endpoint
# ---------------------------------------------------------------------------

output "iscsi_dns_name" {
  description = "The iSCSI endpoint DNS name."
  value       = try(aws_fsx_ontap_storage_virtual_machine.this.endpoints[0].iscsi[0].dns_name, "")
}

output "iscsi_ip_addresses" {
  description = "The iSCSI endpoint IP addresses."
  value       = try(aws_fsx_ontap_storage_virtual_machine.this.endpoints[0].iscsi[0].ip_addresses, [])
}

# ---------------------------------------------------------------------------
# Management Endpoint
# ---------------------------------------------------------------------------

output "management_dns_name" {
  description = "The management endpoint DNS name for ONTAP CLI/REST API access."
  value       = try(aws_fsx_ontap_storage_virtual_machine.this.endpoints[0].management[0].dns_name, "")
}

output "management_ip_addresses" {
  description = "The management endpoint IP addresses."
  value       = try(aws_fsx_ontap_storage_virtual_machine.this.endpoints[0].management[0].ip_addresses, [])
}

# ---------------------------------------------------------------------------
# NFS Endpoint
# ---------------------------------------------------------------------------

output "nfs_dns_name" {
  description = "The NFS endpoint DNS name for mounting volumes."
  value       = try(aws_fsx_ontap_storage_virtual_machine.this.endpoints[0].nfs[0].dns_name, "")
}

output "nfs_ip_addresses" {
  description = "The NFS endpoint IP addresses."
  value       = try(aws_fsx_ontap_storage_virtual_machine.this.endpoints[0].nfs[0].ip_addresses, [])
}

# ---------------------------------------------------------------------------
# SMB Endpoint (only populated when Active Directory is configured)
# ---------------------------------------------------------------------------

output "smb_dns_name" {
  description = "The SMB endpoint DNS name (only available with AD). Use for UNC paths."
  value       = try(aws_fsx_ontap_storage_virtual_machine.this.endpoints[0].smb[0].dns_name, "")
}

output "smb_ip_addresses" {
  description = "The SMB endpoint IP addresses (only populated with AD)."
  value       = try(aws_fsx_ontap_storage_virtual_machine.this.endpoints[0].smb[0].ip_addresses, [])
}
