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
  description = "OpenStackLoadBalancerListenerSpec defines the configuration for an Octavia listener"
  type = object({
    # (Required) The load balancer to attach this listener to.
    # Supports StringValueOrRef pattern - use {value: "lb-id"} for literal values.
    loadbalancer_id = object({
      value = string
    })

    # (Required) The protocol the listener accepts (HTTP, HTTPS, TCP, UDP, TERMINATED_HTTPS).
    protocol = string

    # (Required) The port on which the listener accepts traffic (1-65535).
    protocol_port = number

    # (Optional) Human-readable description of the listener.
    description = optional(string, "")

    # (Optional) Maximum number of connections. -1 means unlimited.
    connection_limit = optional(number, null)

    # (Optional) URI of the Barbican TLS secret container for TLS termination.
    # Required when protocol is TERMINATED_HTTPS.
    default_tls_container_ref = optional(string, "")

    # (Optional) Headers to insert into HTTP requests before forwarding to backends.
    insert_headers = optional(map(string), {})

    # (Optional) List of CIDRs allowed to access this listener.
    allowed_cidrs = optional(list(string), [])

    # (Optional) Administrative state of the listener. Default: true.
    admin_state_up = optional(bool, true)

    # (Optional) Tags applied to the listener in OpenStack.
    tags = optional(list(string), [])

    # (Optional) Override the region from the provider config.
    region = optional(string, "")
  })
}
