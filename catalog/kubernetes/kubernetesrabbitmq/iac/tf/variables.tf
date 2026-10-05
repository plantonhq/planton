# Typed mirror of KubernetesRabbitMqSpec (spec.proto). The spec arrives
# from the proto->tfvars converter in snake_case with every StringValueOrRef
# foreign key -- `namespace` (KubernetesNamespace), `storage_class`
# (KubernetesStorageClass), `tls.secret_name` (KubernetesCertificate secret
# output) and `tls.ca_secret_name` (KubernetesSecret) -- resolved to a
# literal string before Terraform runs. Enum fields arrive as the proto
# enum value names (e.g. "load_balancer", "prefer_dual_stack").
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
  description = "KubernetesRabbitMq specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    replicas         = optional(number, 1)
    image = optional(object({
      repo             = optional(string, "")
      tag              = optional(string, "")
      pull_secret_name = optional(string, "")
    }))
    image_pull_secrets = optional(list(string), [])
    service = optional(object({
      type             = optional(string, "")
      annotations      = optional(map(string), {})
      labels           = optional(map(string), {})
      ip_family_policy = optional(string, "")
    }))
    disk_size     = optional(string, "10Gi")
    storage_class = optional(string, "")
    ephemeral     = optional(bool, false)
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
    configuration = optional(object({
      additional_plugins = optional(list(string), [])
      additional_config  = optional(string, "")
      advanced_config    = optional(string, "")
      env_config         = optional(string, "")
      erlang_inet_config = optional(string, "")
    }))
    tls = optional(object({
      secret_name               = string
      ca_secret_name            = optional(string, "")
      disable_non_tls_listeners = optional(bool, false)
    }))
    tolerations = optional(list(object({
      key                = optional(string, "")
      operator           = optional(string, "")
      value              = optional(string, "")
      effect             = optional(string, "")
      toleration_seconds = optional(number)
    })), [])
    spread_across_nodes              = optional(bool, false)
    termination_grace_period_seconds = optional(number, 604800)
    delay_start_seconds              = optional(number, 30)
    skip_post_deploy_steps           = optional(bool, false)
    auto_enable_all_feature_flags    = optional(bool, false)
    secret_backend = optional(object({
      vault = optional(object({
        role              = string
        default_user_path = string
        annotations       = optional(map(string), {})
        pki_issuer_path   = optional(string, "")
      }))
      external_secret_name = optional(string, "")
    }))
    node_selector = optional(map(string), {})
  })
}
