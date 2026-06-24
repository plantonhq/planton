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
  description = "Specification for Cloudflare Load Balancer"
  type = object({
    # The DNS hostname for the load balancer (e.g., "app.example.com")
    hostname = string

    # Foreign key reference to a Cloudflare DNS zone (StringValueOrRef flattened
    # to a plain string by the tfvars converter).
    zone_id = optional(string)

    # List of origin servers behind this load balancer
    origins = list(object({
      name    = string
      address = string
      weight  = optional(number, 1)
    }))

    # Whether Cloudflare's proxy is enabled for this hostname (orange cloud)
    proxied = optional(bool, true)

    # HTTP path to use for health monitoring of origins
    health_probe_path = optional(string, "/")

    # Session affinity setting ("none" or "cookie"); enum flattens to its string name.
    session_affinity = optional(string, "none")

    # Traffic steering policy ("off"/failover, "geo", or "random"); enum flattens to its string name.
    steering_policy = optional(string, "off")
  })
}
