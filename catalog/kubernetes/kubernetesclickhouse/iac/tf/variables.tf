# Typed mirror of KubernetesClickHouseSpec (spec.proto). The spec arrives
# from the proto->tfvars converter in snake_case with every StringValueOrRef
# foreign key -- `namespace` (KubernetesNamespace), `storage_class` and the
# managed-Keeper `coordination.keeper.storage_class`
# (KubernetesStorageClass), each user's `password` -- resolved to a literal
# string before Terraform runs. Enum fields arrive as the proto enum value
# names (e.g. "managed_keeper", "external_zookeeper").
#
# optional() defaults mirror the proto's (dev.planton.shared.options.default)
# annotations, so the module renders the same resource whether or not the
# platform's defaulting middleware ran.

variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "KubernetesClickHouse specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    version          = string

    image = optional(object({
      repo             = optional(string, "")
      tag              = optional(string, "")
      pull_secret_name = optional(string, "")
    }))

    cluster_name = optional(string, "main")
    shards       = optional(number, 1)
    replicas     = optional(number, 1)

    resources = optional(object({
      limits = optional(object({
        cpu    = optional(string, "")
        memory = optional(string, "")
      }))
      requests = optional(object({
        cpu    = optional(string, "")
        memory = optional(string, "")
      }))
    }))

    disk_size                = string
    storage_class            = optional(string, "")
    log_disk_size            = optional(string, "")
    retain_volumes_on_delete = optional(bool, false)

    coordination = optional(object({
      type = optional(string, "")
      keeper = optional(object({
        replicas = optional(number, 3)
        resources = optional(object({
          limits = optional(object({
            cpu    = optional(string, "")
            memory = optional(string, "")
          }))
          requests = optional(object({
            cpu    = optional(string, "")
            memory = optional(string, "")
          }))
        }))
        disk_size     = optional(string, "10Gi")
        storage_class = optional(string, "")
      }))
      external = optional(object({
        nodes = optional(list(object({
          host = string
          port = optional(number, 2181)
        })), [])
        root     = optional(string, "")
        identity = optional(string, "")
      }))
    }))

    users = optional(list(object({
      name              = string
      password          = string
      profile           = optional(string, "")
      quota             = optional(string, "")
      networks          = optional(list(string), [])
      grants            = optional(list(string), [])
      access_management = optional(bool, false)
      settings          = optional(map(string), {})
    })), [])

    profiles = optional(list(object({
      name     = string
      settings = optional(map(string), {})
    })), [])

    quotas = optional(list(object({
      name     = string
      settings = optional(map(string), {})
    })), [])

    settings = optional(map(string), {})
    files    = optional(map(string), {})

    auto_inter_node_secret       = optional(bool, true)
    spread_replicas_across_nodes = optional(bool, false)
    pdb_max_unavailable          = optional(number, 1)

    node_selector = optional(map(string), {})
    tolerations = optional(list(object({
      key                = optional(string, "")
      operator           = optional(string, "")
      value              = optional(string, "")
      effect             = optional(string, "")
      toleration_seconds = optional(number)
    })), [])

    service_annotations = optional(map(string), {})
    stopped             = optional(bool, false)
    image_pull_secrets  = optional(list(string), [])
  })
}
