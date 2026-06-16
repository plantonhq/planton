variable "spec" {
  description = "AwsFsxOntapVolumeSpec — desired configuration passed from the manifest."
  type        = any
}

# ---------------------------------------------------------------------------
# Resource Metadata
# ---------------------------------------------------------------------------

variable "resource_name" {
  description = "Name for the volume (used for tags and Terraform resource naming)."
  type        = string
  default     = "awsfsxontapvol"
}

variable "labels" {
  description = "Additional labels to merge into resource tags."
  type        = map(string)
  default     = {}
}

# ---------------------------------------------------------------------------
# Parent Reference
# ---------------------------------------------------------------------------

variable "storage_virtual_machine_id" {
  description = "The ID of the Storage Virtual Machine this volume belongs to. Required. ForceNew."
  type        = string
}

# ---------------------------------------------------------------------------
# Volume Identity
# ---------------------------------------------------------------------------

variable "volume_name" {
  description = "ONTAP volume name. 1-203 alphanumeric + underscore only. Required. ForceNew."
  type        = string

  validation {
    condition     = can(regex("^[a-zA-Z0-9_]+$", var.volume_name)) && length(var.volume_name) >= 1 && length(var.volume_name) <= 203
    error_message = "volume_name must be 1-203 alphanumeric characters and underscores only."
  }
}

# ---------------------------------------------------------------------------
# Size
# ---------------------------------------------------------------------------

variable "size_in_megabytes" {
  description = "Volume size in megabytes. Minimum 20 MB."
  type        = number

  validation {
    condition     = var.size_in_megabytes >= 20
    error_message = "size_in_megabytes must be at least 20."
  }
}

# ---------------------------------------------------------------------------
# Volume Configuration
# ---------------------------------------------------------------------------

variable "junction_path" {
  description = "Mount point in SVM namespace (e.g., /vol1). Must start with /."
  type        = string
  default     = ""
}

variable "ontap_volume_type" {
  description = "ONTAP volume type: RW (read-write) or DP (data protection). ForceNew. Default: RW."
  type        = string
  default     = "RW"

  validation {
    condition     = contains(["RW", "DP"], var.ontap_volume_type)
    error_message = "ontap_volume_type must be RW or DP."
  }
}

variable "volume_style" {
  description = "Volume style: FLEXVOL or FLEXGROUP. ForceNew. Default: FLEXVOL."
  type        = string
  default     = "FLEXVOL"

  validation {
    condition     = contains(["FLEXVOL", "FLEXGROUP"], var.volume_style)
    error_message = "volume_style must be FLEXVOL or FLEXGROUP."
  }
}

variable "security_style" {
  description = "Security style: UNIX, NTFS, or MIXED. Inherits from SVM if empty."
  type        = string
  default     = ""
}

variable "snapshot_policy" {
  description = "ONTAP snapshot policy name (e.g., default, none)."
  type        = string
  default     = ""
}

variable "storage_efficiency_enabled" {
  description = "Enable ONTAP deduplication, compression, and compaction."
  type        = bool
  default     = false
}

variable "copy_tags_to_backups" {
  description = "Copy resource tags to automatic volume backups."
  type        = bool
  default     = false
}

# ---------------------------------------------------------------------------
# Deletion Behavior
# ---------------------------------------------------------------------------

variable "skip_final_backup" {
  description = "Skip the automatic backup when the volume is deleted."
  type        = bool
  default     = false
}

variable "bypass_snaplock_enterprise_retention" {
  description = "Allow deletion of SnapLock Enterprise volumes with unexpired WORM files."
  type        = bool
  default     = false
}

# ---------------------------------------------------------------------------
# Tiering Policy (optional)
# ---------------------------------------------------------------------------

variable "tiering_policy" {
  description = "Data tiering policy. Set to null to use default."
  type = object({
    name           = string
    cooling_period = optional(number, 0)
  })
  default = null
}

# ---------------------------------------------------------------------------
# SnapLock Configuration (optional)
# ---------------------------------------------------------------------------

variable "snaplock_configuration" {
  description = "SnapLock WORM configuration. Set to null to disable."
  type = object({
    snaplock_type              = string
    audit_log_volume           = optional(bool, false)
    privileged_delete          = optional(string, "DISABLED")
    volume_append_mode_enabled = optional(bool, false)
    autocommit_period = optional(object({
      type  = string
      value = optional(number, 0)
    }), null)
    retention_period = optional(object({
      default_retention = optional(object({
        type  = string
        value = optional(number, 0)
      }), null)
      minimum_retention = optional(object({
        type  = string
        value = optional(number, 0)
      }), null)
      maximum_retention = optional(object({
        type  = string
        value = optional(number, 0)
      }), null)
    }), null)
  })
  default = null
}

# ---------------------------------------------------------------------------
# Aggregate Configuration (optional — for FLEXGROUP volumes)
# ---------------------------------------------------------------------------

variable "aggregate_configuration" {
  description = "Aggregate configuration for FLEXGROUP volumes. Set to null for FLEXVOL."
  type = object({
    aggregates                 = list(string)
    constituents_per_aggregate = optional(number, 8)
  })
  default = null
}
