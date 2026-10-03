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
  description = "AwsEcrRegistrySettings specification"
  type = object({
    region = string
    registry_policy = optional(string, "")
    scanning = optional(object({
      scan_type = optional(string, "")
      rules = optional(list(object({
        scan_frequency = optional(string, "")
        filters = list(string)
      })), [])
    }))
    replication_rules = optional(list(object({
      destinations = list(object({
        region = string
        registry_id = optional(string, "")
      }))
      repository_filters = optional(list(string), [])
    })), [])
    pull_through_cache_rules = optional(list(object({
      ecr_repository_prefix = string
      upstream_registry_url = string
      upstream_repository_prefix = optional(string, "")
      credential_arn = optional(string, "")
      custom_role_arn = optional(string, "")
    })), [])
    repository_creation_templates = optional(list(object({
      prefix = string
      description = optional(string, "")
      applied_for = list(string)
      custom_role_arn = optional(string, "")
      image_tag_mutability = optional(string, "")
      image_tag_mutability_exclusion_filters = optional(list(string), [])
      encryption = optional(object({
        type = optional(string, "")
        kms_key = optional(string, "")
      }))
      lifecycle_policy = optional(string, "")
      repository_policy = optional(string, "")
      resource_tags = optional(map(string), {})
    })), [])
    account_settings = optional(object({
      basic_scan_type_version = optional(string, "")
      blob_mounting = optional(bool)
      registry_policy_scope = optional(string, "")
    }))
    pull_time_update_exclusions = optional(list(string), [])
  })
}