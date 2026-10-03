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
  description = "AwsSagemakerFeatureGroup specification"
  type = object({
    region = string
    record_identifier_feature_name = string
    event_time_feature_name = string
    description = optional(string, "")
    role_arn = string
    feature_definitions = list(object({
      name = string
      type = optional(string, "")
      collection_type = optional(string, "")
      vector_dimension = optional(number)
    }))
    online_store = optional(object({
      enabled = optional(bool, false)
      kms_key_arn = optional(string, "")
      storage_type = optional(string, "")
      ttl = optional(object({
        unit = optional(string, "")
        value = optional(number, 0)
      }))
    }))
    offline_store = optional(object({
      s3_uri = string
      kms_key_arn = optional(string, "")
      disable_glue_table_creation = optional(bool, false)
      table_format = optional(string, "")
      data_catalog = optional(object({
        catalog = string
        database = string
        table_name = string
      }))
    }))
    throughput = optional(object({
      mode = optional(string, "")
      provisioned_read_capacity_units = optional(number)
      provisioned_write_capacity_units = optional(number)
    }))
  })
}