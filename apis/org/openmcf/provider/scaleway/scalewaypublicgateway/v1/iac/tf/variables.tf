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
  description = "Scaleway Public Gateway specification"
  type = object({
    # Private Network ID (resolved from StringValueOrRef before Terraform runs)
    private_network_id = string

    # Zone where the gateway will be created (e.g., "fr-par-1")
    zone = string

    # Gateway type: "VPC-GW-S" (standard) or "VPC-GW-XL" (high-bandwidth)
    type = string

    # Enable NAT masquerade on the Private Network attachment
    enable_masquerade = optional(bool, true)

    # SSH bastion configuration
    bastion = optional(object({
      enabled           = optional(bool, false)
      port              = optional(number, 22)
      allowed_ip_ranges = optional(list(string), [])
    }))

    # Enable outbound SMTP (port 25)
    enable_smtp = optional(bool, false)

    # Reverse DNS hostname for the public IP
    reverse_dns = optional(string, "")

    # Port forwarding (PAT) rules
    pat_rules = optional(list(object({
      private_ip   = string
      private_port = number
      public_port  = number
      protocol     = optional(string, "both")
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
