# Typed mirror of KubernetesGatekeeperSpec (spec.proto).
# The spec arrives from the proto->tfvars converter in snake_case;
# StringValueOrRef fields arrive resolved to their literal string values.

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
  description = "KubernetesGatekeeper specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string, "3.23.0")
    replicas         = optional(number)
    validating_webhook = optional(object({
      enabled                     = optional(bool)
      failure_policy              = optional(string, "")
      timeout_seconds             = optional(number)
      enable_delete_operations    = optional(bool, false)
      check_ignore_failure_policy = optional(string, "")
    }))
    mutating_webhook = optional(object({
      enabled              = optional(bool)
      failure_policy       = optional(string, "")
      timeout_seconds      = optional(number)
      mutation_annotations = optional(bool, false)
    }))
    audit = optional(object({
      interval_seconds            = optional(number)
      constraint_violations_limit = optional(number)
      from_cache                  = optional(bool, false)
      match_kind_only             = optional(bool, false)
      chunk_size                  = optional(number)
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
    }))
    exempt_namespaces         = optional(list(string), [])
    exempt_namespace_prefixes = optional(list(string), [])
    engine = optional(object({
      enable_external_data                = optional(bool)
      enable_k8s_native_validation        = optional(bool)
      enable_generator_resource_expansion = optional(bool)
      disabled_builtins                   = optional(list(string), [])
      log_denies                          = optional(bool, false)
      log_level                           = optional(string, "")
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
    scheduling = optional(object({
      node_selector = optional(map(string), {})
      tolerations = optional(list(object({
        key                = optional(string, "")
        operator           = optional(string, "")
        value              = optional(string, "")
        effect             = optional(string, "")
        toleration_seconds = optional(number)
      })), [])
    }))
    external_cert = optional(object({
      secret_name = string
    }))
    hooks = optional(object({
      label_namespace                            = optional(bool)
      probe_webhook                              = optional(bool)
      upgrade_crds                               = optional(bool)
      delete_webhook_configurations_on_uninstall = optional(bool, false)
    }))
    image = optional(object({
      repo             = optional(string, "")
      tag              = optional(string, "")
      pull_secret_name = optional(string, "")
    }))
    helm_values = optional(string, "")
  })
}
