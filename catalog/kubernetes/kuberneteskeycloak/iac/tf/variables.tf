# Typed mirror of KubernetesKeycloakSpec (spec.proto). The spec arrives
# from the proto->tfvars converter in snake_case with every
# StringValueOrRef foreign key — `namespace` (KubernetesNamespace),
# `db.host` (KubernetesPostgres rw Service), the secret selectors' `name`
# (KubernetesPostgres app credential Secret) and `http.tls_secret_name`
# (KubernetesCertificate) — resolved to a literal string before Terraform
# runs.
#
# optional() defaults deliberately stay null for the operator-defaulted
# scalars (ports, probe thresholds, strict, update strategy): the module
# renders a CR key only when the spec declares a value, so the operator's
# defaulting stays authoritative — except network_policy_enabled /
# service_monitor_enabled, which locals.tf resolves to their effective
# values because the CR renders them explicitly.

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
  description = "KubernetesKeycloak specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    instances        = optional(number)
    image            = optional(string, "")
    start_optimized  = optional(bool, false)
    db = object({
      vendor   = string
      host     = optional(string, "")
      port     = optional(number)
      database = optional(string, "")
      username_secret = optional(object({
        name = string
        key  = string
      }))
      password_secret = optional(object({
        name = string
        key  = string
      }))
      schema        = optional(string, "")
      jdbc_url      = optional(string, "")
      pool_min_size = optional(number)
      pool_max_size = optional(number)
    })
    http = object({
      tls_secret_name = optional(string, "")
      http_enabled    = optional(bool, false)
      http_port       = optional(number)
      https_port      = optional(number)
    })
    hostname = object({
      hostname            = optional(string, "")
      admin               = optional(string, "")
      strict              = optional(bool)
      backchannel_dynamic = optional(bool, false)
    })
    proxy_headers = optional(string, "")
    features = optional(object({
      enabled  = optional(list(string), [])
      disabled = optional(list(string), [])
    }))
    transaction_xa_enabled = optional(bool, false)
    cache_config = optional(object({
      config_map_name = string
      key             = optional(string)
    }))
    truststore_secret_names = optional(list(string), [])
    additional_options = optional(list(object({
      name  = string
      value = optional(string, "")
      secret = optional(object({
        name = string
        key  = string
      }))
    })), [])
    bootstrap_admin_secret_name = optional(string, "")
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
    scheduling = optional(object({
      node_selector = optional(map(string), {})
      tolerations = optional(list(object({
        key                = optional(string, "")
        operator           = optional(string, "")
        value              = optional(string, "")
        effect             = optional(string, "")
        toleration_seconds = optional(number)
      })), [])
      priority_class_name = optional(string, "")
    }))
    probes = optional(object({
      liveness_failure_threshold  = optional(number)
      liveness_period_seconds     = optional(number)
      readiness_failure_threshold = optional(number)
      readiness_period_seconds    = optional(number)
      startup_failure_threshold   = optional(number)
      startup_period_seconds      = optional(number)
    }))
    http_management_port    = optional(number)
    network_policy_enabled  = optional(bool)
    service_monitor_enabled = optional(bool)
    update = optional(object({
      strategy = optional(string)
      revision = optional(string, "")
    }))
    tracing = optional(object({
      enabled       = optional(bool, false)
      endpoint      = optional(string, "")
      protocol      = optional(string)
      sampler_ratio = optional(string)
    }))
  })
}
