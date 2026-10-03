variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "AwsS3TableBucket specification"
  type = object({
    region = string
    encryption = optional(object({
      sse_algorithm = optional(string, "")
      kms_key_arn = optional(string, "")
    }))
    unreferenced_file_removal = optional(object({
      disabled = optional(bool, false)
      non_current_days = optional(number, 0)
      unreferenced_days = optional(number, 0)
    }))
    force_destroy = optional(bool, false)
    resource_policy = optional(string, "")
    replication = optional(object({
      role = string
      destination_table_bucket_arns = list(string)
    }))
    namespaces = optional(list(object({
      name = string
      tables = optional(list(object({
        name = string
        iceberg_schema = optional(object({
          fields = list(object({
            name = string
            type = string
            required = optional(bool, false)
          }))
        }))
        properties = optional(map(string), {})
        encryption = optional(object({
          sse_algorithm = optional(string, "")
          kms_key_arn = optional(string, "")
        }))
        compaction = optional(object({
          disabled = optional(bool, false)
          target_file_size_mb = optional(number, 0)
        }))
        snapshot_management = optional(object({
          disabled = optional(bool, false)
          max_snapshot_age_hours = optional(number, 0)
          min_snapshots_to_keep = optional(number, 0)
        }))
        resource_policy = optional(string, "")
        replication = optional(object({
          role = string
          destination_table_bucket_arns = list(string)
        }))
      })), [])
    })), [])
  })
}