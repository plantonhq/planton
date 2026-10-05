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
  description = "KubernetesGrafana specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string)
    replicas         = optional(number)
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
    admin_secret = optional(object({
      name         = string
      user_key     = optional(string)
      password_key = optional(string)
    }))
    storage = optional(object({
      size          = optional(string)
      storage_class = optional(string, "")
    }))
    database = optional(object({
      engine = string
      host   = string
      name   = string
      user   = string
      password_secret = object({
        name = string
        key  = string
      })
      ssl_mode = optional(string, "")
    }))
    datasources = optional(list(object({
      name       = string
      type       = optional(string)
      url        = string
      is_default = optional(bool, false)
      uid        = optional(string, "")
      basic_auth = optional(object({
        username = string
        password_secret = object({
          name = string
          key  = string
        })
      }))
      json_data = optional(string, "")
    })), [])
    dashboard_sidecar_enabled = optional(bool)
    community_dashboards = optional(list(object({
      gnet_id    = number
      revision   = optional(number, 0)
      datasource = string
    })), [])
    plugins = optional(list(string), [])
    server = optional(object({
      root_url = optional(string, "")
    }))
    auth = optional(object({
      anonymous_enabled  = optional(bool, false)
      anonymous_org_role = optional(string)
      disable_login_form = optional(bool, false)
      # client_secret arrives resolved from its managed-secret reference.
      google = optional(object({
        client_id             = string
        client_secret         = string
        allowed_domains       = optional(list(string), [])
        hosted_domain         = optional(string, "")
        allow_sign_up         = optional(bool)
        auto_login            = optional(bool, false)
        role_attribute_path   = optional(string, "")
        role_attribute_strict = optional(bool, false)
      }))
      generic_oauth = optional(object({
        name                  = optional(string)
        client_id             = string
        client_secret         = string
        auth_url              = string
        token_url             = string
        api_url               = string
        scopes                = optional(list(string), [])
        email_attribute_path  = optional(string, "")
        login_attribute_path  = optional(string, "")
        name_attribute_path   = optional(string, "")
        role_attribute_path   = optional(string, "")
        role_attribute_strict = optional(bool, false)
        groups_attribute_path = optional(string, "")
        allowed_groups        = optional(list(string), [])
        allowed_domains       = optional(list(string), [])
        allow_sign_up         = optional(bool)
        auto_login            = optional(bool, false)
        use_pkce              = optional(bool)
      }))
    }))
    smtp = optional(object({
      host                    = string
      from_address            = optional(string, "")
      from_name               = optional(string, "")
      credentials_secret_name = optional(string, "")
      skip_verify             = optional(bool, false)
    }))
    service_monitor_enabled = optional(bool, false)
    image = optional(object({
      repository       = optional(string, "")
      tag              = optional(string, "")
      pull_secret_name = optional(string, "")
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
    helm_values = optional(string, "")
  })
}
