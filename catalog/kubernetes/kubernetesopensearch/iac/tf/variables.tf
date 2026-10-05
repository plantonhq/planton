# Typed mirror of KubernetesOpenSearchSpec (spec.proto). The spec arrives
# from the proto->tfvars converter in snake_case with every StringValueOrRef
# foreign key -- `namespace` (KubernetesNamespace), pool
# `persistence.pvc.storage_class` (KubernetesStorageClass), the TLS secrets
# (KubernetesCertificate/KubernetesSecret), the security-config secrets, the
# keystore secrets, the dashboards/monitoring credential secrets -- resolved
# to a literal string before Terraform runs.
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
  description = "KubernetesOpenSearch specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    version          = string
    http_port        = optional(number, 9200)

    node_pools = list(object({
      name     = string
      replicas = number
      roles    = list(string)
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
      jvm       = optional(string, "")
      disk_size = optional(string, "")
      persistence = optional(object({
        pvc = optional(object({
          storage_class = optional(string, "")
        }))
        empty_dir = optional(object({
          size_limit = optional(string, "")
        }))
      }))
      additional_config = optional(map(string), {})
      node_selector     = optional(map(string), {})
      tolerations = optional(list(object({
        key                = optional(string, "")
        operator           = optional(string, "")
        value              = optional(string, "")
        effect             = optional(string, "")
        toleration_seconds = optional(number)
      })), [])
      pdb = optional(object({
        enable          = optional(bool, false)
        min_available   = optional(string, "")
        max_unavailable = optional(string, "")
      }))
    }))

    additional_config    = optional(map(string), {})
    service_annotations  = optional(map(string), {})
    set_vm_max_map_count = optional(bool, true)
    drain_data_nodes     = optional(bool, false)
    plugins_list         = optional(list(string), [])

    bootstrap = optional(object({
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
      jvm               = optional(string, "")
      additional_config = optional(map(string), {})
    }))

    security = optional(object({
      transport_tls = optional(object({
        generate  = optional(bool, true)
        per_node  = optional(bool, true)
        secret    = optional(string, "")
        ca_secret = optional(string, "")
        nodes_dn  = optional(list(string), [])
        admin_dn  = optional(list(string), [])
      }))
      http_tls = optional(object({
        generate = optional(bool, true)
        secret   = optional(string, "")
      }))
      config = optional(object({
        security_config_secret   = optional(string, "")
        admin_secret             = optional(string, "")
        admin_credentials_secret = optional(string, "")
      }))
    }))

    dashboards = optional(object({
      enabled  = optional(bool, false)
      replicas = optional(number, 1)
      version  = optional(string, "")
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
      tls = optional(object({
        enable   = optional(bool, false)
        generate = optional(bool, true)
        secret   = optional(string, "")
      }))
      base_path                     = optional(string, "")
      additional_config             = optional(map(string), {})
      opensearch_credentials_secret = optional(string, "")
      service = optional(object({
        type                        = optional(string, "ClusterIP")
        load_balancer_source_ranges = optional(list(string), [])
      }))
      plugins_list = optional(list(string), [])
    }))

    monitoring = optional(object({
      enabled                = optional(bool, false)
      scrape_interval        = optional(string, "")
      monitoring_user_secret = optional(string, "")
      plugin_url             = optional(string, "")
    }))

    keystore = optional(list(object({
      secret       = string
      key_mappings = optional(map(string), {})
    })), [])

    snapshot_repositories = optional(list(object({
      name     = string
      type     = string
      settings = optional(map(string), {})
    })), [])

    additional_volumes = optional(list(object({
      name            = string
      path            = string
      sub_path        = optional(string, "")
      secret_name     = optional(string, "")
      config_map_name = optional(string, "")
      restart_pods    = optional(bool, false)
    })), [])

    image = optional(object({
      repo             = optional(string, "")
      tag              = optional(string, "")
      pull_secret_name = optional(string, "")
    }))

    image_pull_secrets = optional(list(string), [])
  })
}
