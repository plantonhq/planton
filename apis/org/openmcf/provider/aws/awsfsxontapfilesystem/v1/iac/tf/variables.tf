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
  description = "AWS region where the FSx ONTAP file system will be created."
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
  description = "Name for the FSx ONTAP file system (used for tags)."
  type        = string
  default     = "awsfsxontapfilesystem"
}

variable "labels" {
  description = "Additional labels to merge into resource tags."
  type        = map(string)
  default     = {}
}

# ---------------------------------------------------------------------------
# File System Core (AwsFsxOntapFileSystemSpec)
# ---------------------------------------------------------------------------

variable "deployment_type" {
  description = "Deployment type: SINGLE_AZ_1, SINGLE_AZ_2, MULTI_AZ_1, or MULTI_AZ_2. ONTAP supports scale-out HA pairs for single-AZ. ForceNew."
  type        = string
  default     = "SINGLE_AZ_2"
}

variable "storage_capacity_gib" {
  description = "Storage capacity in GiB. Range: 1024-1048576. ONTAP supports NFS, SMB, and iSCSI with compression and deduplication."
  type        = number
  default     = 1024

  validation {
    condition     = var.storage_capacity_gib >= 1024
    error_message = "storage_capacity_gib must be at least 1024."
  }
}

variable "storage_type" {
  description = "Storage media: SSD (sub-millisecond latency) or HDD (throughput-oriented with SSD cache). ForceNew."
  type        = string
  default     = "SSD"
}

variable "throughput_capacity_per_ha_pair" {
  description = "Throughput per HA pair in MB/s. Total throughput = this value × ha_pairs. Valid: 128, 256, 384, 512, 768, 1024, 1536, 2048, 3072, 4096, 6144."
  type        = number

  validation {
    condition     = var.throughput_capacity_per_ha_pair > 0
    error_message = "throughput_capacity_per_ha_pair must be greater than 0."
  }
}

variable "ha_pairs" {
  description = "Number of HA pairs. Single-AZ: 1-12 for scale-out. Multi-AZ: must be 1. Default: 1."
  type        = number
  default     = 1
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

variable "preferred_subnet_id" {
  description = "Preferred subnet for active file server. MULTI_AZ only. ForceNew."
  type        = string
  default     = ""
}

variable "security_group_ids" {
  description = "Security group IDs. Must allow NFS (111, 2049), SMB (445), iSCSI (3260), ONTAP REST (443). Up to 50."
  type        = list(string)
  default     = []
}

variable "endpoint_ip_address_range" {
  description = "CIDR range for endpoint IPs. MULTI_AZ only. ForceNew."
  type        = string
  default     = ""
}

variable "route_table_ids" {
  description = "Route table IDs for file system routes. MULTI_AZ only. Up to 50."
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
# ONTAP Administration
# ---------------------------------------------------------------------------

variable "fsx_admin_password" {
  description = "ONTAP administrative password for fsxadmin (SSH, REST API). 8-50 characters. Sensitive."
  type        = string
  default     = ""
  sensitive   = true
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
