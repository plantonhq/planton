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
  description = "KubernetesSolrOperator specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string)
    replicas         = optional(number)
    watch_namespaces = optional(list(string), [])
    zookeeper_operator = optional(object({
      install      = optional(bool)
      use_existing = optional(bool, false)
    }))
    leader_election_enabled = optional(bool)
    metrics_enabled         = optional(bool)
    mtls = optional(object({
      client_cert_secret   = string
      ca_cert_secret       = optional(string, "")
      ca_cert_secret_key   = optional(string)
      insecure_skip_verify = optional(bool)
      watch_for_updates    = optional(bool)
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
    node_selector = optional(map(string), {})
    tolerations = optional(list(object({
      key                = optional(string, "")
      operator           = optional(string, "")
      value              = optional(string, "")
      effect             = optional(string, "")
      toleration_seconds = optional(number)
    })), [])
    image_pull_secret = optional(string, "")
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