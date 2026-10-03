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
  description = "AwsSagemakerMlflowServer specification"
  type = object({
    region = string
    artifact_store_uri = string
    role_arn = string
    size = optional(string, "")
    mlflow_version = optional(string, "")
    automatic_model_registration = optional(bool, false)
    weekly_maintenance_window_start = optional(string, "")
  })
}