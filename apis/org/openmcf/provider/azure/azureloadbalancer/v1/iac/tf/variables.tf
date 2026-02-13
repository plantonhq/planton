variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Azure Load Balancer specification"
  type = object({
    # The Azure region where the Load Balancer will be created
    region = string

    # The Azure Resource Group name
    resource_group = string

    # The name of the Load Balancer
    name = string

    # Public IP address ID for public LB (mutually exclusive with subnet_id)
    public_ip_id = optional(string)

    # Subnet ID for internal LB (mutually exclusive with public_ip_id)
    subnet_id = optional(string)

    # Optional static private IP for internal LB
    private_ip_address = optional(string)

    # Backend address pools
    backend_pools = list(object({
      name = string
    }))

    # Health probes
    health_probes = list(object({
      # Probe name
      name = string

      # Protocol: "Tcp", "Http", "Https"
      protocol = string

      # Port to probe (1-65535)
      port = number

      # Request path for Http/Https probes
      request_path = optional(string)

      # Interval between probes in seconds (min 5, default 15)
      interval_in_seconds = optional(number, 15)

      # Consecutive failures before unhealthy (min 1, default 2)
      number_of_probes = optional(number, 2)
    }))

    # Load balancing rules
    rules = list(object({
      # Rule name
      name = string

      # Protocol: "Tcp", "Udp", "All"
      protocol = string

      # Frontend port (0-65534)
      frontend_port = number

      # Backend port (0-65535)
      backend_port = number

      # Name of the backend pool to route to
      backend_pool_name = string

      # Name of the health probe to use
      probe_name = string

      # TCP idle timeout in minutes (4-100, default 4)
      idle_timeout_in_minutes = optional(number, 4)

      # Enable floating IP / Direct Server Return
      enable_floating_ip = optional(bool, false)

      # Disable outbound SNAT for this rule
      disable_outbound_snat = optional(bool, false)
    }))
  })
}
