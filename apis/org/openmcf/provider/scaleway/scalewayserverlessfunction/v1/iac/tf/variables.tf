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
  description = "Scaleway Serverless Function specification"
  type = object({
    # Region for the function namespace and function.
    region = string

    # Language runtime (e.g., "node20", "python311", "go122").
    runtime = string

    # Function handler/entrypoint (e.g., "handler.handle").
    handler = string

    # Privacy: "public" or "private".
    privacy = string

    # Description of the function.
    description = optional(string, "")

    # Memory allocated in MB.
    memory_limit_mb = optional(number, 256)

    # Minimum always-running instances (0 = scale to zero).
    min_scale = optional(number, 0)

    # Maximum concurrent instances.
    max_scale = optional(number, 20)

    # Maximum execution time in seconds.
    timeout_seconds = optional(number, 300)

    # HTTP/HTTPS behavior: "enabled" or "redirected".
    http_option = optional(string, "enabled")

    # Environment variables and secrets.
    env = optional(object({
      variables = optional(list(object({
        name  = string
        value = string
      })), [])
      secrets = optional(list(object({
        name  = string
        value = string
      })), [])
    }), { variables = [], secrets = [] })

    # Private Network ID for VPC connectivity.
    # StringValueOrRef -- use .value for the resolved string.
    private_network_id = optional(object({
      value      = optional(string)
      value_from = optional(object({
        kind       = optional(string)
        env        = optional(string)
        name       = optional(string)
        field_path = optional(string)
      }))
    }))

    # Execution environment (e.g., "v1", "v2").
    sandbox = optional(string, "")

    # Path to zip file containing function source code.
    zip_file = optional(string, "")

    # Hash of zip file for change detection.
    zip_hash = optional(string, "")

    # Scheduled CRON triggers.
    cron_triggers = optional(list(object({
      name     = optional(string, "")
      schedule = string
      args     = string
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
