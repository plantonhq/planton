variable "metadata" {
  description = "Resource metadata (name, org, env, id, labels)"
  type = object({
    name = string
    org  = optional(string, "")
    env  = optional(string, "")
    id   = optional(string, "")
  })
}

variable "spec" {
  description = "AwsGlueCatalogDatabase spec"
  type = object({
    description  = optional(string, "")
    location_uri = optional(string, "")
  })
}
