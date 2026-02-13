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
  description = "Scaleway Serverless Container specification"
  type = object({
    # Region for the container namespace and container.
    region = string

    # Container image configuration.
    image = object({
      # Registry endpoint URL (pre-resolved from StringValueOrRef).
      registry_endpoint = string
      # Image name within the registry.
      name = string
      # Image tag.
      tag = string
    })

    # Deployment trigger string (SHA256 digest or any change indicator).
    registry_sha256 = optional(string, "")

    # Listening port exposed by the container.
    port = optional(number, 8080)

    # Privacy: "public" or "private".
    privacy = string

    # Description of the container.
    description = optional(string, "")

    # Memory allocated in MB.
    memory_limit_mb = optional(number, 256)

    # vCPU in milliCPU (0 = auto from memory).
    cpu_limit = optional(number, 0)

    # Minimum always-running instances (0 = scale to zero).
    min_scale = optional(number, 0)

    # Maximum concurrent instances.
    max_scale = optional(number, 20)

    # Maximum request time in seconds.
    timeout_seconds = optional(number, 300)

    # HTTP/HTTPS behavior: "enabled" or "redirected".
    http_option = optional(string, "enabled")

    # Communication protocol: "http1" or "h2c".
    protocol = optional(string, "http1")

    # CMD override.
    commands = optional(list(string), [])

    # ENTRYPOINT args override.
    args = optional(list(string), [])

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
    # StringValueOrRef -- Terraform receives the pre-resolved string.
    private_network_id = optional(string)

    # Execution environment (e.g., "v1", "v2").
    sandbox = optional(string, "")

    # HTTP health check configuration.
    health_check = optional(object({
      path               = string
      failure_threshold  = optional(number, 3)
      interval_seconds   = optional(number, 30)
    }))

    # Autoscaling thresholds.
    scaling_option = optional(object({
      concurrent_requests_threshold = optional(number, 0)
      cpu_usage_threshold           = optional(number, 0)
      memory_usage_threshold        = optional(number, 0)
    }))

    # Local ephemeral storage in MB.
    local_storage_limit_mb = optional(number, 0)

    # Whether to deploy (start) the container.
    deploy = optional(bool, true)

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
