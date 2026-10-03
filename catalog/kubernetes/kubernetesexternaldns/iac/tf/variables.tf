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
  description = "KubernetesExternalDns specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string)
    aws_route53 = optional(object({
      region                  = optional(string, "")
      zone_id_filters         = optional(list(string), [])
      zone_type               = optional(string)
      assume_role             = optional(string, "")
      assume_role_external_id = optional(string, "")
      access_key_id           = optional(string, "")
      secret_access_key       = optional(string, "")
    }))
    google_cloud_dns = optional(object({
      project                  = string
      zone_id_filters          = optional(list(string), [])
      zone_visibility          = optional(string)
      service_account_key_json = optional(string, "")
    }))
    azure_dns = optional(object({
      resource_group             = string
      subscription_id            = string
      tenant_id                  = optional(string, "")
      private_zones              = optional(bool, false)
      zone_id_filters            = optional(list(string), [])
      managed_identity_client_id = optional(string, "")
      client_id                  = optional(string, "")
      client_secret              = optional(string, "")
    }))
    cloudflare = optional(object({
      api_token            = string
      zone_id_filters      = optional(list(string), [])
      proxied              = optional(bool, false)
      dns_records_per_page = optional(number)
    }))
    webhook = optional(object({
      image_repository = string
      image_tag        = optional(string, "")
      args             = optional(list(string), [])
      env              = optional(map(string), {})
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
    in_memory = optional(object({
      zones = optional(list(string), [])
    }))
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
    sources               = optional(list(string), [])
    gateway_listener_sets = optional(bool, false)
    policy                = optional(string)
    registry              = optional(string)
    txt_owner_id          = optional(string, "")
    txt_prefix            = optional(string, "")
    txt_suffix            = optional(string, "")
    dynamodb_table        = optional(string, "")
    dynamodb_region       = optional(string, "")
    domain_filters        = optional(list(string), [])
    exclude_domains       = optional(list(string), [])
    annotation_filter     = optional(string, "")
    label_filter          = optional(string, "")
    managed_record_types  = optional(list(string), [])
    interval              = optional(string)
    trigger_loop_on_event = optional(bool, false)
    namespaced            = optional(bool, false)
    log_level             = optional(string)
    log_format            = optional(string)
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
    priority_class_name = optional(string, "")
    prometheus = optional(object({
      service_monitor          = optional(bool, false)
      service_monitor_interval = optional(string, "")
      service_monitor_labels   = optional(map(string), {})
    }))
    image_repository = optional(string, "")
    image_tag        = optional(string, "")
    helm_values      = optional(string, "")
  })
}
