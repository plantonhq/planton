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
  description = "Scaleway Instance Security Group specification"
  type = object({
    # Zone where the security group is created (e.g., "fr-par-1")
    zone = string

    # Human-readable description
    description = optional(string, "")

    # Whether the security group is stateful (default: true)
    stateful = optional(bool, true)

    # Default policy for inbound traffic ("accept" or "drop", default: "accept")
    inbound_default_policy = optional(string, "accept")

    # Default policy for outbound traffic ("accept" or "drop", default: "accept")
    outbound_default_policy = optional(string, "accept")

    # Whether to enable default SMTP security (default: true)
    enable_default_security = optional(bool, true)

    # Inbound (ingress) rules
    inbound_rules = optional(list(object({
      action     = string
      protocol   = optional(string, "TCP")
      port_range = optional(string, "")
      ip_range   = optional(string, "")
    })), [])

    # Outbound (egress) rules
    outbound_rules = optional(list(object({
      action     = string
      protocol   = optional(string, "TCP")
      port_range = optional(string, "")
      ip_range   = optional(string, "")
    })), [])
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
