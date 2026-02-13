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
  description = "Scaleway DNS Record specification"
  type = object({
    # DNS zone name where the record will be created.
    # StringValueOrRef -- use .value for the resolved string.
    zone_name = object({
      value      = optional(string)
      value_from = optional(object({
        kind       = optional(string)
        env        = optional(string)
        name       = optional(string)
        field_path = optional(string)
      }))
    })

    # Record name relative to the zone ("" for apex, "www", "api", etc.)
    name = optional(string, "")

    # DNS record type (A, AAAA, ALIAS, CAA, CNAME, DNAME, MX, NS, PTR, SOA, SRV, TXT, TLSA)
    type = string

    # Record data -- literal value or resolved reference.
    # StringValueOrRef -- use .value for the resolved string.
    data = object({
      value      = optional(string)
      value_from = optional(object({
        kind       = optional(string)
        env        = optional(string)
        name       = optional(string)
        field_path = optional(string)
      }))
    })

    # TTL in seconds (default: 3600)
    ttl = optional(number, 3600)

    # Priority for MX/SRV records (default: 0)
    priority = optional(number, 0)

    # Keep zone when this is the last record destroyed (default: true)
    keep_empty_zone = optional(bool, true)
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
