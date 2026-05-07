#########################################
# variables.tf
#
# Defines the input variables for a CronJobKubernetes resource.
#########################################

variable "metadata" {
  description = "Metadata for the CronJob resource, including name, org, env, labels, etc."
  type = object({
    name   = string
    id     = optional(string)
    org    = optional(string)
    env    = optional(string)
    labels = optional(map(string))
    tags   = optional(list(string))
    version = optional(object({
      id      = string
      message = string
    }))
  })
}

variable "spec" {
  description = "Spec defines the configuration for the CronJobKubernetes resource."
  type = object({
    namespace                     = string
    create_namespace              = bool
    schedule                      = string
    concurrency_policy            = optional(string)
    suspend                       = optional(bool)
    successful_jobs_history_limit = optional(number)
    failed_jobs_history_limit     = optional(number)
    backoff_limit                 = optional(number)
    starting_deadline_seconds     = optional(number)
    restart_policy                = optional(string)
    image = object({
      repo             = string
      tag              = string
      pull_secret_name = optional(string)
    })
    resources = object({
      limits = object({
        cpu    = string
        memory = string
      })
      requests = object({
        cpu    = string
        memory = string
      })
    })
    env = optional(object({
      # A map of environment variable names to their string values.
      # StringValueOrRef wrapper types are flattened to plain strings by the
      # tfvars generator before reaching the TF module.
      variables = optional(map(string))
      secrets = optional(map(object({
        value = optional(string)
        secret_ref = optional(object({
          namespace = optional(string)
          name      = string
          key       = string
        }))
      })))
    }))
    command = optional(list(string))
    args    = optional(list(string))

    # ConfigMaps to create alongside the CronJob.
    # Key is the ConfigMap name, value is the content.
    config_maps = optional(map(string), {})

    # Volume mounts for the CronJob container.
    # Supports mounting ConfigMaps, Secrets, HostPaths, EmptyDirs, and PVCs.
    volume_mounts = optional(list(object({
      # Name of the volume mount. Must be unique within the container.
      name = string

      # Path within the container at which the volume should be mounted.
      mount_path = string

      # Whether the volume should be mounted read-only.
      read_only = optional(bool, false)

      # Path within the volume from which the container's volume should be mounted.
      sub_path = optional(string)

      # ConfigMap volume source.
      config_map = optional(object({
        name         = string
        key          = optional(string)
        path         = optional(string)
        default_mode = optional(number)
      }))

      # Secret volume source.
      secret = optional(object({
        name         = string
        key          = optional(string)
        path         = optional(string)
        default_mode = optional(number)
      }))

      # HostPath volume source.
      host_path = optional(object({
        path = string
        type = optional(string)
      }))

      # EmptyDir volume source.
      empty_dir = optional(object({
        medium     = optional(string)
        size_limit = optional(string)
      }))

      # PVC volume source.
      pvc = optional(object({
        claim_name = string
        read_only  = optional(bool, false)
      }))
    })), [])
  })
}

variable "docker_config_json" {
  description = <<EOT
Optional Docker credentials in JSON format to create
an image pull secret (type: kubernetes.io/dockerconfigjson).
Leave empty if no private repo auth is needed.
EOT
  type        = string
  default     = ""
  sensitive   = true
}
