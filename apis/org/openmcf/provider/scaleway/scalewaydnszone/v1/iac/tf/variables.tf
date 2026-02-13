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
  description = "Scaleway DNS Zone specification"
  type = object({
    # The registered parent domain name (e.g., "example.com").
    # Cannot be changed after creation.
    domain = string

    # Subdomain prefix for this zone.
    # Empty string for root zone, or a value like "staging" for
    # subdomain zones (e.g., "staging.example.com").
    subdomain = optional(string, "")

    # Inline DNS records to create within the zone.
    records = optional(list(object({
      # Record name relative to the zone ("" or "@" for apex)
      name = optional(string, "")
      # DNS record type (e.g., "A", "AAAA", "CNAME", "MX", "TXT")
      type = string
      # Record data - literal value or resolved reference
      data = object({
        value = optional(string)
        value_from_resource_output = optional(object({
          resource_id_ref = object({
            name = string
          })
          output_key = string
        }))
      })
      # TTL in seconds (default: 3600)
      ttl = optional(number, 3600)
      # Priority for MX/SRV records (default: 0)
      priority = optional(number, 0)
    })), [])
  })
}

# ── Scaleway Provider Credentials ─────────────────────────────────────

variable "scaleway_access_key" {
  description = "Scaleway API access key"
  type        = string
  sensitive   = true
}

variable "scaleway_secret_key" {
  description = "Scaleway API secret key"
  type        = string
  sensitive   = true
}

variable "scaleway_project_id" {
  description = "Scaleway project ID"
  type        = string
  default     = ""
}

variable "scaleway_organization_id" {
  description = "Scaleway organization ID"
  type        = string
  default     = ""
}
