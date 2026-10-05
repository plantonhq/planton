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
  description = "KubernetesKafka specification"
  type = object({
    namespace = string
    create_namespace = optional(bool, false)
    kafka_version = optional(string, "")
    metadata_version = optional(string, "")
    node_pools = list(object({
      name = string
      roles = list(string)
      replicas = number
      storage = object({
        type = optional(string)
        size = optional(string, "")
        storage_class = optional(string, "")
        delete_claim = optional(bool, false)
        volumes = optional(list(object({
          id = optional(number, 0)
          size = string
          storage_class = optional(string, "")
          delete_claim = optional(bool, false)
          kraft_metadata = optional(bool, false)
        })), [])
      })
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
    }))
    listeners = list(object({
      name = string
      port = number
      type = optional(string)
      tls = optional(bool, false)
      authentication = optional(object({
        type = string
        sasl = optional(bool, false)
        listener_config = optional(map(string), {})
      }))
      configuration = optional(object({
        class = optional(string, "")
        external_traffic_policy = optional(string)
        load_balancer_source_ranges = optional(list(string), [])
        allocate_load_balancer_node_ports = optional(bool)
        create_bootstrap_service = optional(bool)
        use_service_dns_domain = optional(bool, false)
        max_connections = optional(number)
        max_connection_creation_rate = optional(number)
        preferred_node_port_address_type = optional(string)
        publish_not_ready_addresses = optional(bool, false)
        broker_cert_chain_and_key = optional(object({
          secret_name = string
          certificate = optional(string)
          key = optional(string)
        }))
        bootstrap = optional(object({
          host = optional(string, "")
          annotations = optional(map(string), {})
          labels = optional(map(string), {})
          load_balancer_ip = optional(string, "")
          node_port = optional(number)
          alternative_names = optional(list(string), [])
        }))
        brokers = optional(list(object({
          broker = optional(number, 0)
          host = optional(string, "")
          advertised_host = optional(string, "")
          advertised_port = optional(number)
          annotations = optional(map(string), {})
          labels = optional(map(string), {})
          load_balancer_ip = optional(string, "")
          node_port = optional(number)
        })), [])
      }))
    }))
    config = optional(map(string), {})
    authorization = optional(object({
      type = string
      super_users = optional(list(string), [])
      authorizer_class = optional(string, "")
      supports_admin_api = optional(bool, false)
    }))
    entity_operator = optional(object({
      topic_operator_enabled = optional(bool)
      user_operator_enabled = optional(bool)
    }))
    cruise_control = optional(object({
      enabled = optional(bool, false)
      config = optional(map(string), {})
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
      auto_rebalance_modes = optional(list(string), [])
    }))
    kafka_exporter = optional(object({
      enabled = optional(bool, false)
      group_regex = optional(string, "")
      topic_regex = optional(string, "")
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
    }))
    metrics = optional(object({
      enabled = optional(bool, false)
    }))
    cluster_ca = optional(object({
      validity_days = optional(number)
      renewal_days = optional(number)
    }))
    clients_ca = optional(object({
      validity_days = optional(number)
      renewal_days = optional(number)
    }))
    rack = optional(object({
      topology_key = string
    }))
    jvm = optional(object({
      xms = optional(string, "")
      xmx = optional(string, "")
    }))
    maintenance_time_windows = optional(list(string), [])
  })
}
