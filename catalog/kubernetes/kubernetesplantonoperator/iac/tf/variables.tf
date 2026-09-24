variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "KubernetesPlantonOperator specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string, "")
    chart_repository = optional(string, "")
    crds = optional(object({
      install           = optional(bool)
      keep_on_uninstall = optional(bool)
    }))
    replicas        = optional(number)
    leader_election = optional(bool)
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
    service_account = optional(object({
      create      = optional(bool)
      name        = optional(string, "")
      annotations = optional(map(string), {})
    }))
    common_labels   = optional(map(string), {})
    pod_annotations = optional(map(string), {})
    node_selector   = optional(map(string), {})
    tolerations = optional(list(object({
      key                = optional(string, "")
      operator           = optional(string, "")
      value              = optional(string, "")
      effect             = optional(string, "")
      toleration_seconds = optional(number)
    })), [])
    image_pull_secrets = optional(list(string), [])
    image = optional(object({
      repository = optional(string, "")
      tag        = optional(string, "")
    }))
    helm_values = optional(string, "")
  })
}