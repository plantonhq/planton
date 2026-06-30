##############################################
# variables.tf
#
# Input variables for KubernetesGhaRunnerScaleSet
##############################################

variable "metadata" {
  description = "Cloud resource metadata (name, id, org, env)"
  type = object({
    name = string
    id   = optional(string, "")
    org  = optional(string, "")
    env  = optional(string, "")
  })
}

variable "spec" {
  description = "KubernetesGhaRunnerScaleSet specification"
  type = object({
    namespace = object({
      value = string
    })
    create_namespace    = optional(bool, false)
    helm_chart_version  = optional(string, "0.13.1")
    
    github = object({
      config_url = string
      pat_token = optional(object({
        token = string
      }))
      github_app = optional(object({
        app_id             = string
        installation_id    = string
        private_key_base64 = string  # Base64 encoded PEM format
      }))
      existing_secret_name = optional(string)
    })
    
    scaling = optional(object({
      min_runners = optional(number, 0)
      max_runners = optional(number, 5)
    }), {})
    
    runner_group          = optional(string, "")
    runner_scale_set_name = optional(string, "")
    
    container_mode = object({
      type = string
      work_volume_claim = optional(object({
        storage_class = optional(string, "")
        size          = string
        access_modes  = optional(list(string), ["ReadWriteOnce"])
      }))
    })
    
    runner = optional(object({
      image = optional(object({
        repository  = optional(string, "")
        tag         = optional(string, "")
        pull_policy = optional(string, "")
      }))
      resources = optional(object({
        requests = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
        limits = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
      }))
      env = optional(list(object({
        name  = string
        value = string
      })), [])
      volume_mounts = optional(list(object({
        name       = string
        mount_path = string
        read_only  = optional(bool, false)
        sub_path   = optional(string, "")
      })), [])
    }))
    
    persistent_volumes = optional(list(object({
      name          = string
      storage_class = optional(string, "")
      size          = string
      access_modes  = optional(list(string), ["ReadWriteOnce"])
      mount_path    = string
      read_only     = optional(bool, false)
    })), [])
    
    controller_service_account = optional(object({
      namespace = optional(string, "")
      name      = optional(string, "")
    }))
    
    image_pull_secrets = optional(list(string), [])
    labels             = optional(map(string), {})
    annotations        = optional(map(string), {})
  })
}

