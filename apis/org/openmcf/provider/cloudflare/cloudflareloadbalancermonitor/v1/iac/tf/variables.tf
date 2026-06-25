variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "CloudflareLoadBalancerMonitorSpec defines an account-scoped health monitor"
  # NOTE: the type enum flattens to its string name ("http", "https", ...) via the
  # proto->tfvars converter; "monitor_type_unspecified"/"" both mean the default (http).
  type = object({
    # (Required) The Cloudflare account ID that owns this monitor.
    account_id = string

    # (Optional) Health-check protocol. Defaults to http.
    type = optional(string, "")

    description    = optional(string, "")
    path           = optional(string, "")
    expected_codes = optional(string, "")
    expected_body  = optional(string, "")
    method         = optional(string, "")
    probe_zone     = optional(string, "")

    # HTTP request headers (name -> one or more values).
    headers = optional(list(object({
      name   = string
      values = list(string)
    })), [])

    # 0 means "use the Cloudflare default" for the numeric tuning knobs.
    port             = optional(number, 0)
    interval         = optional(number, 0)
    timeout          = optional(number, 0)
    retries          = optional(number, 0)
    consecutive_up   = optional(number, 0)
    consecutive_down = optional(number, 0)

    follow_redirects = optional(bool, false)
    allow_insecure   = optional(bool, false)
  })
}
