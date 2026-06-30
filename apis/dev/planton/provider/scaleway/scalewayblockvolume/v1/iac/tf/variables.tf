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
  description = "Scaleway Block Volume specification"
  type = object({
    zone             = string
    size_gb          = number
    performance_tier = string
    snapshot_id      = optional(string, "")
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
