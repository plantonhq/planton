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
  description = "CloudflareHealthcheck specification"
  type = object({
    zone_id = string
    name = string
    address = string
    type = optional(string)
    check_regions = optional(list(string), [])
    consecutive_fails = optional(number)
    consecutive_successes = optional(number)
    interval = optional(number)
    retries = optional(number)
    timeout = optional(number)
    suspended = optional(bool)
    http_config = optional(object({
      method = optional(string)
      path = optional(string)
      port = optional(number)
      expected_codes = optional(list(string), [])
      expected_body = optional(string, "")
      follow_redirects = optional(bool)
      allow_insecure = optional(bool)
      headers = optional(map(object({
        values = list(string)
      })), {})
    }))
    tcp_config = optional(object({
      method = optional(string)
      port = optional(number)
    }))
  })
}