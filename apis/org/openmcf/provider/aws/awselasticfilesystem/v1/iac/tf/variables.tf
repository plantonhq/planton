# ---------------------------------------------------------------------------
# Resource Metadata
# ---------------------------------------------------------------------------

variable "resource_name" {
  description = "Name for the EFS file system (used for tags and creation_token)."
  type        = string
  default     = "awselasticfilesystem"
}

variable "labels" {
  description = "Additional labels to merge into resource tags."
  type        = map(string)
  default     = {}
}

# ---------------------------------------------------------------------------
# File System Core (AwsElasticFileSystemSpec)
# ---------------------------------------------------------------------------

variable "encrypted" {
  description = "Enable encryption at rest. ForceNew — cannot be changed after creation."
  type        = bool
  default     = true
}

variable "kms_key_id" {
  description = "Customer-managed KMS key ARN for encryption. ForceNew. Requires encrypted = true."
  type        = string
  default     = ""
}

variable "performance_mode" {
  description = "Performance mode: generalPurpose or maxIO. ForceNew. Default: generalPurpose."
  type        = string
  default     = "generalPurpose"
}

variable "throughput_mode" {
  description = "Throughput mode: bursting, provisioned, or elastic. Default: bursting."
  type        = string
  default     = "bursting"
}

variable "provisioned_throughput_in_mibps" {
  description = "Provisioned throughput in MiB/s. Only valid when throughput_mode = provisioned."
  type        = number
  default     = 0
}

variable "availability_zone_name" {
  description = "AZ name for One Zone storage (e.g., us-east-1a). ForceNew."
  type        = string
  default     = ""
}

# ---------------------------------------------------------------------------
# Lifecycle Policies
# ---------------------------------------------------------------------------

variable "transition_to_ia" {
  description = "Transition to IA storage. Valid: AFTER_1_DAY, AFTER_7_DAYS, etc."
  type        = string
  default     = ""
}

variable "transition_to_archive" {
  description = "Transition IA files to Archive. Requires transition_to_ia."
  type        = string
  default     = ""
}

variable "transition_to_primary_storage_class" {
  description = "Transition back to Standard on access. Valid: AFTER_1_ACCESS."
  type        = string
  default     = ""
}

# ---------------------------------------------------------------------------
# Backup
# ---------------------------------------------------------------------------

variable "backup_enabled" {
  description = "Enable automatic daily backups via AWS Backup."
  type        = bool
  default     = false
}

# ---------------------------------------------------------------------------
# Mount Targets (Networking)
# ---------------------------------------------------------------------------

variable "subnet_ids" {
  description = "Subnet IDs for mount targets. One mount target per subnet. Required."
  type        = list(string)
}

variable "security_group_ids" {
  description = "Security group IDs for mount targets. Must allow NFS (TCP 2049)."
  type        = list(string)
  default     = []
}

# ---------------------------------------------------------------------------
# Access Points
# ---------------------------------------------------------------------------

variable "access_points" {
  description = "Access points with name, posix_user, and root_directory."
  type = list(object({
    name = string
    posix_user = optional(object({
      uid            = number
      gid            = number
      secondary_gids = optional(list(number))
    }))
    root_directory = optional(object({
      path = string
      creation_info = optional(object({
        owner_uid   = number
        owner_gid   = number
        permissions = string
      }))
    }))
  }))
  default = []
}

# ---------------------------------------------------------------------------
# Resource Policy
# ---------------------------------------------------------------------------

variable "policy" {
  description = "JSON IAM resource policy for the file system."
  type        = string
  default     = ""
}
