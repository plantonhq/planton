variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "AwsSesAccountSettings specification"
  type = object({
    region = string
    suppression = optional(object({
      reasons = optional(list(string), [])
      enabled = optional(bool)
    }))
    vdm = optional(object({
      # Required by the API inside the block; no literal default so an unset
      # value can never be mistaken for an explicit false.
      enabled                   = optional(bool)
      engagement_metrics        = optional(bool)
      optimized_shared_delivery = optional(bool)
    }))
  })
}
