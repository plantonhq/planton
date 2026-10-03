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
  description = "DigitalOceanDatabaseKafkaTopic specification"
  type = object({
    cluster = string
    topic_name = string
    partition_count = optional(number)
    replication_factor = optional(number)
    config = optional(object({
      cleanup_policy = optional(string, "")
      compression_type = optional(string, "")
      delete_retention_ms = optional(number)
      file_delete_delay_ms = optional(number)
      flush_messages = optional(number)
      flush_ms = optional(number)
      index_interval_bytes = optional(number)
      max_compaction_lag_ms = optional(number)
      max_message_bytes = optional(number)
      message_down_conversion_enable = optional(bool)
      message_format_version = optional(string, "")
      message_timestamp_difference_max_ms = optional(number)
      message_timestamp_type = optional(string, "")
      min_cleanable_dirty_ratio = optional(number)
      min_compaction_lag_ms = optional(number)
      min_insync_replicas = optional(number)
      preallocate = optional(bool)
      retention_bytes = optional(number)
      retention_ms = optional(number)
      segment_bytes = optional(number)
      segment_index_bytes = optional(number)
      segment_jitter_ms = optional(number)
      segment_ms = optional(number)
    }))
  })
}