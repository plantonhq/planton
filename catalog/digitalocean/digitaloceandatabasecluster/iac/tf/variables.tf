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
  description = "DigitalOceanDatabaseCluster specification"
  type = object({
    cluster_name = string
    engine = string
    engine_version = string
    region = string
    size_slug = string
    node_count = number
    vpc = optional(string, "")
    storage_gib = optional(number, 0)
    maintenance_window = optional(object({
      day = string
      hour = string
    }))
    backup_restore = optional(object({
      database_name = string
      backup_created_at = optional(string, "")
    }))
    storage_autoscale = optional(object({
      enabled = bool
      threshold_percent = optional(number, 0)
      increment_gib = optional(number, 0)
    }))
    eviction_policy = optional(string, "")
    sql_mode = optional(string, "")
    project_id = optional(string, "")
    tags = optional(list(string), [])
  })
}
