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
  description = "OpenStackLoadBalancerMonitorSpec defines the configuration for an Octavia health monitor"
  type = object({
    # (Required) The pool to monitor.
    # Supports StringValueOrRef pattern - use {value: "pool-id"} for literal values.
    pool_id = object({
      value = string
    })

    # (Required) The type of health check to perform.
    # Valid values: HTTP, HTTPS, PING, TCP, TLS-HELLO, UDP-CONNECT
    type = string

    # (Required) The interval in seconds between health checks.
    delay = number

    # (Required) The maximum time in seconds to wait for a response.
    timeout = number

    # (Required) The number of consecutive successful checks before a member is healthy.
    # Must be between 1 and 10.
    max_retries = number

    # (Optional) The number of consecutive failed checks before a member is unhealthy.
    # Must be between 1 and 10. Default: same as max_retries.
    max_retries_down = optional(number)

    # (Optional) URL path for HTTP/HTTPS health checks.
    url_path = optional(string, "")

    # (Optional) HTTP method for HTTP/HTTPS health checks.
    http_method = optional(string, "")

    # (Optional) Expected HTTP response codes for a healthy member.
    expected_codes = optional(string, "")

    # (Optional) Administrative state of the monitor. Default: true.
    admin_state_up = optional(bool, true)

    # (Optional) Override the region from the provider config.
    region = optional(string, "")

    # Note: Health monitors do NOT support tags in the Terraform provider.
  })
}
