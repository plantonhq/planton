variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "DigitalOceanContainerRegistry specification"
  type = object({
    name              = string
    subscription_tier = string
    region            = optional(string, "")
    docker_credentials = optional(object({
      write          = optional(bool, false)
      expiry_seconds = optional(number)
    }))
  })
}