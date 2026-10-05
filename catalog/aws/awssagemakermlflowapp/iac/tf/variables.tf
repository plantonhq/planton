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
  description = "AwsSagemakerMlflowApp specification"
  type = object({
    region = string
    artifact_store_uri = string
    role_arn = string
    account_default_status = optional(string, "")
    default_domain_ids = optional(list(string), [])
    model_registration_mode = optional(string, "")
    weekly_maintenance_window_start = optional(string, "")
  })
}