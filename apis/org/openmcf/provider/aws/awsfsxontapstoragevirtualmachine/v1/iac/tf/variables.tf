variable "spec" {
  description = "AwsFsxOntapStorageVirtualMachineSpec — desired configuration passed from the manifest."
  type        = any
}

# ---------------------------------------------------------------------------
# Resource Metadata
# ---------------------------------------------------------------------------

variable "resource_name" {
  description = "Name for the SVM (used for tags and Terraform resource naming)."
  type        = string
  default     = "awsfsxontapsvm"
}

variable "labels" {
  description = "Additional labels to merge into resource tags."
  type        = map(string)
  default     = {}
}

# ---------------------------------------------------------------------------
# Parent Reference
# ---------------------------------------------------------------------------

variable "file_system_id" {
  description = "The ID of the FSx for ONTAP file system this SVM belongs to. Required. ForceNew."
  type        = string

  validation {
    condition     = length(var.file_system_id) >= 11
    error_message = "file_system_id must be at least 11 characters (e.g., fs-0123456789a)."
  }
}

# ---------------------------------------------------------------------------
# SVM Identity
# ---------------------------------------------------------------------------

variable "svm_name" {
  description = "ONTAP SVM name. 1-47 alphanumeric + underscore only. Required. ForceNew."
  type        = string

  validation {
    condition     = can(regex("^[a-zA-Z0-9_]+$", var.svm_name)) && length(var.svm_name) >= 1 && length(var.svm_name) <= 47
    error_message = "svm_name must be 1-47 alphanumeric characters and underscores only."
  }
}

# ---------------------------------------------------------------------------
# Volume Security
# ---------------------------------------------------------------------------

variable "root_volume_security_style" {
  description = "Security style for root volume: UNIX, NTFS, or MIXED. ForceNew. Default: UNIX."
  type        = string
  default     = "UNIX"

  validation {
    condition     = contains(["UNIX", "NTFS", "MIXED"], var.root_volume_security_style)
    error_message = "root_volume_security_style must be UNIX, NTFS, or MIXED."
  }
}

# ---------------------------------------------------------------------------
# SVM Administration
# ---------------------------------------------------------------------------

variable "svm_admin_password" {
  description = "Password for vsadmin (SVM-scoped SSH/REST). 8-50 characters. Sensitive."
  type        = string
  default     = ""
  sensitive   = true
}

# ---------------------------------------------------------------------------
# Active Directory Configuration (optional — for SMB access)
# ---------------------------------------------------------------------------

variable "active_directory_configuration" {
  description = "Active Directory configuration for SMB access. Set to null to disable."
  type = object({
    netbios_name                         = optional(string, "")
    domain_name                          = string
    dns_ips                              = list(string)
    username                             = string
    password                             = string
    file_system_administrators_group     = optional(string, "Domain Admins")
    organizational_unit_distinguished_name = optional(string, "")
  })
  default   = null
  sensitive = true
}
