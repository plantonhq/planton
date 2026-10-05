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
  description = "KubernetesCertManager specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string)
    crds = optional(object({
      install           = optional(bool)
      keep_on_uninstall = optional(bool)
    }))
    replicas = optional(number)
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
    log_level                    = optional(number)
    cluster_resource_namespace   = optional(string, "")
    leader_election_namespace    = optional(string)
    enable_certificate_owner_ref = optional(bool, false)
    feature_gates                = optional(map(bool), {})
    dns01_self_check = optional(object({
      recursive_nameservers      = optional(list(string), [])
      recursive_nameservers_only = optional(bool, false)
    }))
    max_concurrent_challenges = optional(number)
    workload_identity = optional(object({
      gke = optional(object({
        service_account_email = string
      }))
      eks = optional(object({
        role_arn = string
      }))
      aks = optional(object({
        client_id = string
        tenant_id = optional(string)
      }))
    }))
    image_registry = optional(string, "")
    prometheus = optional(object({
      enabled                  = optional(bool)
      service_monitor          = optional(bool, false)
      service_monitor_interval = optional(string)
      service_monitor_labels   = optional(map(string), {})
    }))
    node_selector = optional(map(string), {})
    tolerations = optional(list(object({
      key                = optional(string, "")
      operator           = optional(string, "")
      value              = optional(string, "")
      effect             = optional(string, "")
      toleration_seconds = optional(number)
    })), [])
    pod_disruption_budget = optional(bool, false)
    webhook = optional(object({
      replicas        = optional(number)
      timeout_seconds = optional(number)
      host_network    = optional(bool, false)
      secure_port     = optional(number)
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
    }))
    cainjector = optional(object({
      enabled  = optional(bool)
      replicas = optional(number)
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
    }))
    startupapicheck = optional(object({
      enabled = optional(bool)
      timeout = optional(string)
    }))
    helm_values = optional(string, "")
  })
}