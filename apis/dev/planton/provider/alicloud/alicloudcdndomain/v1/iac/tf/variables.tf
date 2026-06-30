variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name   = string
    id     = optional(string)
    org    = optional(string)
    env    = optional(string)
    labels = optional(map(string))
    tags   = optional(list(string))
  })
}

variable "spec" {
  description = "Alibaba Cloud CDN domain specification"
  type = object({
    region      = string
    domain_name = string
    cdn_type    = string
    scope       = optional(string, "")
    sources = list(object({
      type     = string
      content  = string
      port     = optional(number, 80)
      priority = optional(number, 20)
      weight   = optional(number, 10)
    }))
    certificate_config = optional(object({
      cert_name                   = optional(string, "")
      cert_type                   = optional(string, "")
      cert_id                     = optional(string, "")
      cert_region                 = optional(string, "")
      server_certificate          = optional(string, "")
      private_key                 = optional(string, "")
      server_certificate_status   = optional(string, "on")
    }))
    check_url         = optional(string, "")
    resource_group_id = optional(string, "")
    tags              = optional(map(string), {})
  })

  validation {
    condition     = contains(["web", "download", "video"], var.spec.cdn_type)
    error_message = "cdn_type must be one of: web, download, video."
  }

  validation {
    condition     = var.spec.scope == "" || contains(["domestic", "overseas", "global"], var.spec.scope)
    error_message = "scope must be one of: domestic, overseas, global."
  }

  validation {
    condition     = length(var.spec.sources) >= 1
    error_message = "At least one source is required."
  }
}
