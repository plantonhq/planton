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
  description = "AWS region where the FSx Lustre file system will be created."
  type        = string
}

variable "session_token" {
  description = "AWS session token for temporary credentials (e.g., assumed role)."
  type        = string
  default     = ""
  sensitive   = true
}

variable "spec" {
  description = "AwsFsxLustreFileSystemSpec — desired configuration passed from the manifest."
  type        = any
}

# ---------------------------------------------------------------------------
# Resource Metadata
# ---------------------------------------------------------------------------

variable "resource_name" {
  description = "Name for the FSx Lustre file system (used for tags)."
  type        = string
  default     = "awsfsxlustrefilesystem"
}

variable "labels" {
  description = "Additional labels to merge into resource tags."
  type        = map(string)
  default     = {}
}

# ---------------------------------------------------------------------------
# File System Core (AwsFsxLustreFileSystemSpec)
# ---------------------------------------------------------------------------

variable "deployment_type" {
  description = "Deployment type: SCRATCH_1, SCRATCH_2, PERSISTENT_1, or PERSISTENT_2. ForceNew."
  type        = string
  default     = "SCRATCH_2"
}

variable "storage_capacity_gib" {
  description = "Storage capacity in GiB. Minimum 1200. Valid increments depend on deployment and storage type."
  type        = number
  default     = 1200

  validation {
    condition     = var.storage_capacity_gib >= 1200
    error_message = "storage_capacity_gib must be at least 1200."
  }
}

variable "storage_type" {
  description = "Storage media type: SSD or HDD. HDD only valid for PERSISTENT_1. ForceNew."
  type        = string
  default     = "SSD"
}

variable "per_unit_storage_throughput" {
  description = "Throughput per TiB in MB/s. Required for PERSISTENT_1 and PERSISTENT_2. Invalid for SCRATCH."
  type        = number
  default     = 0
}

variable "data_compression_type" {
  description = "Data compression type: NONE or LZ4. Can be changed after creation."
  type        = string
  default     = "NONE"
}

variable "file_system_type_version" {
  description = "Lustre version (e.g., 2.12, 2.15). ForceNew. Leave empty for latest."
  type        = string
  default     = ""
}

# ---------------------------------------------------------------------------
# Networking (ForceNew)
# ---------------------------------------------------------------------------

variable "subnet_id" {
  description = "Subnet ID for the file system. Lustre is single-AZ — exactly one subnet. Required."
  type        = string
}

variable "security_group_ids" {
  description = "Security group IDs. Must allow TCP 988 and TCP 1018-1023 (Lustre). Up to 50."
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
# S3 Data Repository (Legacy — ForceNew)
# ---------------------------------------------------------------------------

variable "import_path" {
  description = "S3 URI to import data from (e.g., s3://my-bucket/prefix). SCRATCH only. ForceNew."
  type        = string
  default     = ""
}

variable "export_path" {
  description = "S3 URI for exporting data. Requires import_path. ForceNew."
  type        = string
  default     = ""
}

# ---------------------------------------------------------------------------
# Logging
# ---------------------------------------------------------------------------

variable "log_destination" {
  description = "CloudWatch Logs log group ARN for Lustre audit events. Empty disables logging."
  type        = string
  default     = ""
}

variable "log_level" {
  description = "Audit log level: DISABLED, WARN_ONLY, ERROR_ONLY, or WARN_ERROR."
  type        = string
  default     = "WARN_ERROR"
}

# ---------------------------------------------------------------------------
# Backup (PERSISTENT only)
# ---------------------------------------------------------------------------

variable "automatic_backup_retention_days" {
  description = "Days to retain automatic backups (0-90). 0 disables. PERSISTENT only."
  type        = number
  default     = 0
}

variable "daily_automatic_backup_start_time" {
  description = "Daily UTC time to start backups in HH:MM format. PERSISTENT only."
  type        = string
  default     = ""
}

variable "copy_tags_to_backups" {
  description = "Copy file system tags to backups."
  type        = bool
  default     = false
}

variable "skip_final_backup" {
  description = "Skip final backup on deletion. PERSISTENT only."
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

# ---------------------------------------------------------------------------
# Metadata Configuration (PERSISTENT_2 only)
# ---------------------------------------------------------------------------

variable "metadata_mode" {
  description = "Metadata IOPS mode: AUTOMATIC or USER_PROVISIONED. PERSISTENT_2 only."
  type        = string
  default     = ""
}

variable "metadata_iops" {
  description = "Metadata IOPS when mode is USER_PROVISIONED. Ignored in AUTOMATIC."
  type        = number
  default     = 0
}
