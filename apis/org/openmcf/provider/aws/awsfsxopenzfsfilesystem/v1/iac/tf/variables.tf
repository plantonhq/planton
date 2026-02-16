# ---------------------------------------------------------------------------
# Provider Configuration
# ---------------------------------------------------------------------------

variable "access_key" {
  description = "AWS access key ID for provider authentication."
  type        = string
  default     = ""
  sensitive   = true
}

variable "secret_key" {
  description = "AWS secret access key for provider authentication."
  type        = string
  default     = ""
  sensitive   = true
}

variable "region" {
  description = "AWS region where the FSx OpenZFS file system will be created."
  type        = string
}

variable "session_token" {
  description = "AWS session token for temporary credentials (e.g., assumed role)."
  type        = string
  default     = ""
  sensitive   = true
}

# ---------------------------------------------------------------------------
# Resource Metadata
# ---------------------------------------------------------------------------

variable "resource_name" {
  description = "Name for the FSx OpenZFS file system (used for tags)."
  type        = string
  default     = "awsfsxopenzfsfilesystem"
}

variable "labels" {
  description = "Additional labels to merge into resource tags."
  type        = map(string)
  default     = {}
}

# ---------------------------------------------------------------------------
# File System Core (AwsFsxOpenzfsFileSystemSpec)
# ---------------------------------------------------------------------------

variable "deployment_type" {
  description = "Deployment type: SINGLE_AZ_1, SINGLE_AZ_2, or MULTI_AZ_1. ForceNew."
  type        = string
  default     = "SINGLE_AZ_2"
}

variable "storage_capacity_gib" {
  description = "Storage capacity in GiB. Range: 64-524288."
  type        = number
  default     = 64

  validation {
    condition     = var.storage_capacity_gib >= 64
    error_message = "storage_capacity_gib must be at least 64."
  }
}

variable "throughput_capacity" {
  description = "Throughput in MB/s. Valid values depend on deployment type."
  type        = number

  validation {
    condition     = var.throughput_capacity > 0
    error_message = "throughput_capacity must be greater than 0."
  }
}

# ---------------------------------------------------------------------------
# Networking (ForceNew)
# ---------------------------------------------------------------------------

variable "subnet_ids" {
  description = "Subnet IDs. 1 for SINGLE_AZ, 2 for MULTI_AZ. Required."
  type        = list(string)

  validation {
    condition     = length(var.subnet_ids) >= 1
    error_message = "At least one subnet ID is required."
  }
}

variable "security_group_ids" {
  description = "Security group IDs. Must allow TCP 111, 2049, 20001-20003 (NFS). Up to 50."
  type        = list(string)
  default     = []
}

variable "preferred_subnet_id" {
  description = "Preferred subnet for active file server. MULTI_AZ_1 only. ForceNew."
  type        = string
  default     = ""
}

variable "endpoint_ip_address_range" {
  description = "CIDR range for endpoint IPs. MULTI_AZ_1 only. ForceNew."
  type        = string
  default     = ""
}

variable "route_table_ids" {
  description = "Route table IDs for file system routes. MULTI_AZ_1 only. Up to 50."
  type        = list(string)
  default     = []
}

# ---------------------------------------------------------------------------
# Encryption
# ---------------------------------------------------------------------------

variable "kms_key_id" {
  description = "Customer-managed KMS key ARN for encryption at rest. ForceNew. Omit for AWS-managed key."
  type        = string
  default     = ""
}

# ---------------------------------------------------------------------------
# Disk IOPS Configuration
# ---------------------------------------------------------------------------

variable "disk_iops_mode" {
  description = "IOPS mode: AUTOMATIC or USER_PROVISIONED."
  type        = string
  default     = ""
}

variable "disk_iops" {
  description = "Total SSD IOPS. Only when mode is USER_PROVISIONED."
  type        = number
  default     = 0
}

# ---------------------------------------------------------------------------
# Root Volume Configuration
# ---------------------------------------------------------------------------

variable "root_data_compression_type" {
  description = "Root volume compression: NONE, ZSTD, or LZ4."
  type        = string
  default     = "NONE"
}

variable "root_read_only" {
  description = "Root volume read-only flag."
  type        = bool
  default     = false
}

variable "root_record_size_kib" {
  description = "Root volume ZFS record size in KiB: 4, 8, 16, 32, 64, 128, 256, 512, 1024."
  type        = number
  default     = 128
}

variable "root_copy_tags_to_snapshots" {
  description = "Copy root volume tags to snapshots."
  type        = bool
  default     = false
}

variable "root_nfs_client_configurations" {
  description = "NFS client configurations for root volume. List of {clients, options}."
  type = list(object({
    clients = string
    options = list(string)
  }))
  default = []
}

variable "root_user_and_group_quotas" {
  description = "User/group quotas for root volume. List of {id, storage_capacity_quota_gib, type}."
  type = list(object({
    id                        = number
    storage_capacity_quota_gib = number
    type                      = string
  }))
  default = []
}

# ---------------------------------------------------------------------------
# Backup
# ---------------------------------------------------------------------------

variable "automatic_backup_retention_days" {
  description = "Days to retain automatic backups (0-90). 0 disables."
  type        = number
  default     = 0
}

variable "daily_automatic_backup_start_time" {
  description = "Daily UTC time to start backups in HH:MM format."
  type        = string
  default     = ""
}

variable "copy_tags_to_backups" {
  description = "Copy file system tags to backups."
  type        = bool
  default     = false
}

variable "copy_tags_to_volumes" {
  description = "Copy file system tags to volumes."
  type        = bool
  default     = false
}

variable "skip_final_backup" {
  description = "Skip final backup on deletion."
  type        = bool
  default     = true
}

# ---------------------------------------------------------------------------
# Maintenance
# ---------------------------------------------------------------------------

variable "weekly_maintenance_start_time" {
  description = "Weekly UTC maintenance window in d:HH:MM format (1=Mon, 7=Sun)."
  type        = string
  default     = ""
}
