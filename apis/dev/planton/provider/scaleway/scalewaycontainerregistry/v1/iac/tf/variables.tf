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
  description = "Scaleway Container Registry specification"
  type = object({
    # Region where the registry namespace will be created
    # (e.g., "fr-par", "nl-ams", "pl-waw")
    region = string

    # Human-readable description of the registry namespace
    description = optional(string, "")

    # Whether images can be pulled without authentication
    is_public = optional(bool, false)
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
