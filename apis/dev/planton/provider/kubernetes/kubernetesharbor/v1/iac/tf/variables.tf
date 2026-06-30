variable "harbor_kubernetes" {
  description = "Harbor Kubernetes resource configuration"
  type = object({
    metadata = object({
      name = string
      id   = optional(string)
    })
    spec = object({
      namespace        = string
      create_namespace = bool
      core_container = optional(object({
        replicas = number
        resources = optional(object({
          requests = optional(object({
            cpu    = string
            memory = string
          }))
          limits = optional(object({
            cpu    = string
            memory = string
          }))
        }))
      }))
      database = object({
        is_external = bool
        external_database = optional(object({
          host        = string
          port        = number
          username    = string
          password    = string
          use_ssl     = bool
        }))
      })
      cache = object({
        is_external = bool
        external_cache = optional(object({
          host     = string
          port     = number
          password = string
        }))
      })
      storage = object({
        type = string
        s3 = optional(object({
          bucket     = string
          region     = string
          access_key = string
          secret_key = string
        }))
        filesystem = optional(object({
          disk_size = string
        }))
      })
      ingress = optional(object({
        core = optional(object({
          enabled  = bool
          hostname = string
        }))
      }))
      helm_values = optional(map(string))
    })
  })
}

