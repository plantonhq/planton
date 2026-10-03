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
  description = "KubernetesOpenSearchOperator specification"
  type = object({
    namespace                 = string
    create_namespace          = optional(bool, false)
    chart_version             = optional(string)
    watch_namespace           = optional(string, "")
    use_role_bindings         = optional(bool, false)
    log_level                 = optional(string)
    dns_base                  = optional(string)
    parallel_recovery_enabled = optional(bool)
    pprof_endpoints_enabled   = optional(bool, false)
    kube_rbac_proxy_enabled   = optional(bool)
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
    node_selector = optional(map(string), {})
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
    crds = optional(object({
      install           = optional(bool)
      keep_on_uninstall = optional(bool)
    }))
  })
}