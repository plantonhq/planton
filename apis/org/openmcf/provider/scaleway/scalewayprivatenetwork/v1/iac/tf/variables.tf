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
  description = "Scaleway Private Network specification"
  type = object({
    # VPC ID (resolved from StringValueOrRef before Terraform runs)
    vpc_id = string

    # Region where the Private Network will be created
    region = string

    # IPv4 subnet CIDR (e.g., "192.168.0.0/24"). If null, IPAM auto-allocates.
    ipv4_subnet = optional(string)

    # IPv6 subnet CIDRs. Optional for dual-stack networking.
    ipv6_subnets = optional(list(string), [])

    # Whether to propagate default v4/v6 routes
    enable_default_route_propagation = optional(bool, false)
  })
}

variable "scaleway_access_key" {
  description = "Scaleway access key for API authentication"
  type        = string
  sensitive   = true
}

variable "scaleway_secret_key" {
  description = "Scaleway secret key for API authentication"
  type        = string
  sensitive   = true
}

variable "scaleway_project_id" {
  description = "Scaleway project ID (optional, defaults from provider)"
  type        = string
  default     = ""
}

variable "scaleway_organization_id" {
  description = "Scaleway organization ID (optional, defaults from provider)"
  type        = string
  default     = ""
}
