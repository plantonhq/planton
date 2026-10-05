# Typed mirror of KubernetesSolrSpec (spec.proto). The spec arrives from
# the proto->tfvars converter in snake_case with every StringValueOrRef
# foreign key -- `namespace` (KubernetesNamespace), the two
# `storage_class` references (KubernetesStorageClass), and
# `security.basic_auth_secret` (KubernetesSecret) -- resolved to a literal
# string before Terraform runs.
#
# optional() defaults mirror the proto's (dev.planton.shared.options.default)
# annotations, so the module renders the same resource whether or not the
# platform's defaulting middleware ran. Fields whose ABSENCE is meaningful
# (availability.pdb_enabled, the scaling flags) carry NO default: the
# module renders them only when explicitly set, exactly like the Pulumi
# module's presence checks.

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
  description = "KubernetesSolr specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    replicas         = optional(number, 3)
    version          = string
    image_repository = optional(string, "")

    zookeeper = optional(object({
      provided = optional(object({
        replicas = optional(number, 3)
        persistence = optional(object({
          size          = optional(string, "")
          storage_class = optional(string, "")
        }))
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
        chroot = optional(string, "/")
      }))
      external = optional(object({
        connection_string = string
        chroot            = optional(string, "/")
      }))
    }))

    storage = optional(object({
      persistent = optional(object({
        size           = string
        storage_class  = optional(string, "")
        reclaim_policy = optional(string, "Retain")
      }))
      ephemeral = optional(object({
        size_limit = optional(string, "")
      }))
    }))

    java_mem  = optional(string, "")
    solr_opts = optional(string, "")
    log_level = optional(string, "INFO")
    gc_tune   = optional(string, "")

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

    security = optional(object({
      authentication_type = optional(string, "")
      basic_auth_secret   = optional(string, "")
      probes_require_auth = optional(bool, false)
      bootstrap_security_json = optional(object({
        name = string
        key  = string
      }))
    }))

    tls = optional(object({
      pkcs12_secret = object({
        name = string
        key  = string
      })
      keystore_password_secret = object({
        name = string
        key  = string
      })
      truststore_secret = optional(object({
        name = string
        key  = string
      }))
      truststore_password_secret = optional(object({
        name = string
        key  = string
      }))
      client_auth            = optional(string, "None")
      verify_client_hostname = optional(bool, false)
    }))

    backup_repositories = optional(list(object({
      name = string
      s3 = optional(object({
        region        = string
        bucket        = string
        base_location = optional(string, "")
        endpoint      = optional(string, "")
        credentials = optional(object({
          access_key_id_secret = optional(object({
            name = string
            key  = string
          }))
          secret_access_key_secret = optional(object({
            name = string
            key  = string
          }))
        }))
      }))
      gcs = optional(object({
        bucket = string
        gcs_credential_secret = optional(object({
          name = string
          key  = string
        }))
        base_location = optional(string, "")
      }))
      volume = optional(object({
        pvc_claim_name = string
        directory      = optional(string, "")
      }))
    })), [])

    solr_modules    = optional(list(string), [])
    additional_libs = optional(list(string), [])

    update_strategy = optional(object({
      method                         = optional(string, "Managed")
      max_pods_unavailable           = optional(string, "")
      max_shard_replicas_unavailable = optional(string, "")
      restart_schedule               = optional(string, "")
    }))

    availability = optional(object({
      pdb_enabled = optional(bool)
    }))

    scaling = optional(object({
      vacate_pods_on_scale_down = optional(bool)
      populate_pods_on_scale_up = optional(bool)
    }))

    external = optional(object({
      method               = string
      domain_name          = string
      use_external_address = optional(bool, false)
      hide_common          = optional(bool, false)
      hide_nodes           = optional(bool, false)
    }))

    pod_port      = optional(number, 8983)
    node_selector = optional(map(string), {})
    tolerations = optional(list(object({
      key                = optional(string, "")
      operator           = optional(string, "")
      value              = optional(string, "")
      effect             = optional(string, "")
      toleration_seconds = optional(number)
    })), [])
  })
}
