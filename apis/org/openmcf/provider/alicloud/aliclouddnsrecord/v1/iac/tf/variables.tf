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
  description = "Alibaba Cloud DNS record specification"
  type = object({
    region      = string
    domain_name = string
    rr          = string
    type        = string
    value       = string
    ttl         = optional(number, 600)
    priority    = optional(number, 0)
    line        = optional(string, "")
    status      = optional(string, "")
    remark      = optional(string, "")
  })

  validation {
    condition     = length(var.spec.domain_name) >= 1 && length(var.spec.domain_name) <= 253
    error_message = "domain_name must be between 1 and 253 characters."
  }

  validation {
    condition     = contains(["A", "AAAA", "CNAME", "MX", "TXT", "NS", "SRV", "CAA", "REDIRECT_URL", "FORWORD_URL"], var.spec.type)
    error_message = "type must be one of: A, AAAA, CNAME, MX, TXT, NS, SRV, CAA, REDIRECT_URL, FORWORD_URL."
  }

  validation {
    condition     = var.spec.status == "" || contains(["ENABLE", "DISABLE"], var.spec.status)
    error_message = "status must be either ENABLE or DISABLE."
  }
}
