variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "KubernetesStrimziKafkaOperator specification"
  type = object({
    namespace = string
    create_namespace = optional(bool, false)
    chart_version = optional(string)
    replicas = optional(number)
    watch = optional(object({
      any_namespace = optional(bool, false)
      namespaces = optional(list(string), [])
    }))
    full_reconciliation_interval_ms = optional(number)
    operation_timeout_ms = optional(number)
    log_level = optional(string)
    feature_gates = optional(string, "")
    kubernetes_service_dns_domain = optional(string)
    leader_election_enabled = optional(bool)
    generate_network_policy = optional(bool)
    generate_pod_disruption_budget = optional(bool)
    create_global_resources = optional(bool)
    resources = optional(object({
      limits = optional(object({
        cpu = optional(string, "")
        memory = optional(string, "")
      }))
      requests = optional(object({
        cpu = optional(string, "")
        memory = optional(string, "")
      }))
    }))
    node_selector = optional(map(string), {})
    tolerations = optional(list(object({
      key = optional(string, "")
      operator = optional(string, "")
      value = optional(string, "")
      effect = optional(string, "")
      toleration_seconds = optional(number)
    })), [])
    image_pull_secrets = optional(list(string), [])
    image = optional(object({
      registry = optional(string, "")
      repository = optional(string, "")
      tag = optional(string, "")
    }))
    helm_values = optional(string, "")
  })
}
