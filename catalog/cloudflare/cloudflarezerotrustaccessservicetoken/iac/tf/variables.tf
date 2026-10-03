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
  description = "CloudflareZeroTrustAccessServiceToken specification"
  type = object({
    account_id = optional(string, "")
    zone_id = optional(string, "")
    name = string
    duration = optional(string, "")
    client_secret_version = optional(number)
    previous_client_secret_expires_at = optional(string, "")
  })
}