variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Specification for KubernetesDestinationRule"
  type = object({
    # Namespace the DestinationRule is created in (resolved foreign key).
    namespace = string

    # The service registry host this rule applies to. Required. A plain registry reference
    # resolved by istiod (not an OpenMCF foreign key).
    host = string

    # Traffic policy applied across all ports (load balancing, connection pool, outlier
    # detection, TLS, per-port overrides, tunnel, PROXY protocol).
    traffic_policy = optional(object({
      load_balancer = optional(object({
        simple = optional(string)
        consistent_hash = optional(object({
          http_header_name          = optional(string)
          http_cookie               = optional(object({ name = string, path = optional(string), ttl = optional(string) }))
          use_source_ip             = optional(bool)
          http_query_parameter_name = optional(string)
          ring_hash                 = optional(object({ minimum_ring_size = optional(number) }))
          maglev                    = optional(object({ table_size = optional(number) }))
          minimum_ring_size         = optional(number)
        }))
        locality_lb_setting = optional(object({
          distribute        = optional(list(object({ from = optional(string), to = optional(map(number)) })))
          failover          = optional(list(object({ from = optional(string), to = optional(string) })))
          failover_priority = optional(list(string))
          enabled           = optional(bool)
        }))
        warmup_duration_secs = optional(string)
        warmup               = optional(object({ duration = string, minimum_percent = optional(number), aggression = optional(number) }))
      }))
      connection_pool = optional(object({
        tcp = optional(object({
          max_connections         = optional(number)
          connect_timeout         = optional(string)
          tcp_keepalive           = optional(object({ probes = optional(number), time = optional(string), interval = optional(string) }))
          max_connection_duration = optional(string)
          idle_timeout            = optional(string)
        }))
        http = optional(object({
          http1_max_pending_requests  = optional(number)
          http2_max_requests          = optional(number)
          max_requests_per_connection = optional(number)
          max_retries                 = optional(number)
          idle_timeout                = optional(string)
          h2_upgrade_policy           = optional(string)
          use_client_protocol         = optional(bool)
          max_concurrent_streams      = optional(number)
        }))
      }))
      outlier_detection = optional(object({
        split_external_local_origin_errors = optional(bool)
        consecutive_local_origin_failures  = optional(number)
        consecutive_gateway_errors         = optional(number)
        consecutive_5xx_errors             = optional(number)
        interval                           = optional(string)
        base_ejection_time                 = optional(string)
        max_ejection_percent               = optional(number)
        min_health_percent                 = optional(number)
      }))
      tls = optional(object({
        mode                 = optional(string)
        client_certificate   = optional(string)
        private_key          = optional(string)
        ca_certificates      = optional(string)
        credential_name      = optional(string)
        subject_alt_names    = optional(list(string))
        sni                  = optional(string)
        insecure_skip_verify = optional(bool)
        ca_crl               = optional(string)
      }))
      port_level_settings = optional(list(object({
        port = optional(object({ number = number }))
        load_balancer = optional(object({
          simple = optional(string)
          consistent_hash = optional(object({
            http_header_name          = optional(string)
            http_cookie               = optional(object({ name = string, path = optional(string), ttl = optional(string) }))
            use_source_ip             = optional(bool)
            http_query_parameter_name = optional(string)
            ring_hash                 = optional(object({ minimum_ring_size = optional(number) }))
            maglev                    = optional(object({ table_size = optional(number) }))
            minimum_ring_size         = optional(number)
          }))
          locality_lb_setting = optional(object({
            distribute        = optional(list(object({ from = optional(string), to = optional(map(number)) })))
            failover          = optional(list(object({ from = optional(string), to = optional(string) })))
            failover_priority = optional(list(string))
            enabled           = optional(bool)
          }))
          warmup_duration_secs = optional(string)
          warmup               = optional(object({ duration = string, minimum_percent = optional(number), aggression = optional(number) }))
        }))
        connection_pool = optional(object({
          tcp = optional(object({
            max_connections         = optional(number)
            connect_timeout         = optional(string)
            tcp_keepalive           = optional(object({ probes = optional(number), time = optional(string), interval = optional(string) }))
            max_connection_duration = optional(string)
            idle_timeout            = optional(string)
          }))
          http = optional(object({
            http1_max_pending_requests  = optional(number)
            http2_max_requests          = optional(number)
            max_requests_per_connection = optional(number)
            max_retries                 = optional(number)
            idle_timeout                = optional(string)
            h2_upgrade_policy           = optional(string)
            use_client_protocol         = optional(bool)
            max_concurrent_streams      = optional(number)
          }))
        }))
        outlier_detection = optional(object({
          split_external_local_origin_errors = optional(bool)
          consecutive_local_origin_failures  = optional(number)
          consecutive_gateway_errors         = optional(number)
          consecutive_5xx_errors             = optional(number)
          interval                           = optional(string)
          base_ejection_time                 = optional(string)
          max_ejection_percent               = optional(number)
          min_health_percent                 = optional(number)
        }))
        tls = optional(object({
          mode                 = optional(string)
          client_certificate   = optional(string)
          private_key          = optional(string)
          ca_certificates      = optional(string)
          credential_name      = optional(string)
          subject_alt_names    = optional(list(string))
          sni                  = optional(string)
          insecure_skip_verify = optional(bool)
          ca_crl               = optional(string)
        }))
      })))
      tunnel         = optional(object({ protocol = optional(string), target_host = string, target_port = number }))
      proxy_protocol = optional(object({ version = optional(string) }))
    }))

    # Named subsets (e.g. service versions) selected by labels, with optional per-subset
    # traffic-policy overrides (same shape as the top-level traffic_policy).
    subsets = optional(list(object({
      name   = string
      labels = optional(map(string))
      traffic_policy = optional(object({
        load_balancer = optional(object({
          simple = optional(string)
          consistent_hash = optional(object({
            http_header_name          = optional(string)
            http_cookie               = optional(object({ name = string, path = optional(string), ttl = optional(string) }))
            use_source_ip             = optional(bool)
            http_query_parameter_name = optional(string)
            ring_hash                 = optional(object({ minimum_ring_size = optional(number) }))
            maglev                    = optional(object({ table_size = optional(number) }))
            minimum_ring_size         = optional(number)
          }))
          locality_lb_setting = optional(object({
            distribute        = optional(list(object({ from = optional(string), to = optional(map(number)) })))
            failover          = optional(list(object({ from = optional(string), to = optional(string) })))
            failover_priority = optional(list(string))
            enabled           = optional(bool)
          }))
          warmup_duration_secs = optional(string)
          warmup               = optional(object({ duration = string, minimum_percent = optional(number), aggression = optional(number) }))
        }))
        connection_pool = optional(object({
          tcp = optional(object({
            max_connections         = optional(number)
            connect_timeout         = optional(string)
            tcp_keepalive           = optional(object({ probes = optional(number), time = optional(string), interval = optional(string) }))
            max_connection_duration = optional(string)
            idle_timeout            = optional(string)
          }))
          http = optional(object({
            http1_max_pending_requests  = optional(number)
            http2_max_requests          = optional(number)
            max_requests_per_connection = optional(number)
            max_retries                 = optional(number)
            idle_timeout                = optional(string)
            h2_upgrade_policy           = optional(string)
            use_client_protocol         = optional(bool)
            max_concurrent_streams      = optional(number)
          }))
        }))
        outlier_detection = optional(object({
          split_external_local_origin_errors = optional(bool)
          consecutive_local_origin_failures  = optional(number)
          consecutive_gateway_errors         = optional(number)
          consecutive_5xx_errors             = optional(number)
          interval                           = optional(string)
          base_ejection_time                 = optional(string)
          max_ejection_percent               = optional(number)
          min_health_percent                 = optional(number)
        }))
        tls = optional(object({
          mode                 = optional(string)
          client_certificate   = optional(string)
          private_key          = optional(string)
          ca_certificates      = optional(string)
          credential_name      = optional(string)
          subject_alt_names    = optional(list(string))
          sni                  = optional(string)
          insecure_skip_verify = optional(bool)
          ca_crl               = optional(string)
        }))
        port_level_settings = optional(list(object({
          port = optional(object({ number = number }))
          load_balancer = optional(object({
            simple = optional(string)
            consistent_hash = optional(object({
              http_header_name          = optional(string)
              http_cookie               = optional(object({ name = string, path = optional(string), ttl = optional(string) }))
              use_source_ip             = optional(bool)
              http_query_parameter_name = optional(string)
              ring_hash                 = optional(object({ minimum_ring_size = optional(number) }))
              maglev                    = optional(object({ table_size = optional(number) }))
              minimum_ring_size         = optional(number)
            }))
            locality_lb_setting = optional(object({
              distribute        = optional(list(object({ from = optional(string), to = optional(map(number)) })))
              failover          = optional(list(object({ from = optional(string), to = optional(string) })))
              failover_priority = optional(list(string))
              enabled           = optional(bool)
            }))
            warmup_duration_secs = optional(string)
            warmup               = optional(object({ duration = string, minimum_percent = optional(number), aggression = optional(number) }))
          }))
          connection_pool = optional(object({
            tcp = optional(object({
              max_connections         = optional(number)
              connect_timeout         = optional(string)
              tcp_keepalive           = optional(object({ probes = optional(number), time = optional(string), interval = optional(string) }))
              max_connection_duration = optional(string)
              idle_timeout            = optional(string)
            }))
            http = optional(object({
              http1_max_pending_requests  = optional(number)
              http2_max_requests          = optional(number)
              max_requests_per_connection = optional(number)
              max_retries                 = optional(number)
              idle_timeout                = optional(string)
              h2_upgrade_policy           = optional(string)
              use_client_protocol         = optional(bool)
              max_concurrent_streams      = optional(number)
            }))
          }))
          outlier_detection = optional(object({
            split_external_local_origin_errors = optional(bool)
            consecutive_local_origin_failures  = optional(number)
            consecutive_gateway_errors         = optional(number)
            consecutive_5xx_errors             = optional(number)
            interval                           = optional(string)
            base_ejection_time                 = optional(string)
            max_ejection_percent               = optional(number)
            min_health_percent                 = optional(number)
          }))
          tls = optional(object({
            mode                 = optional(string)
            client_certificate   = optional(string)
            private_key          = optional(string)
            ca_certificates      = optional(string)
            credential_name      = optional(string)
            subject_alt_names    = optional(list(string))
            sni                  = optional(string)
            insecure_skip_verify = optional(bool)
            ca_crl               = optional(string)
          }))
        })))
        tunnel         = optional(object({ protocol = optional(string), target_host = string, target_port = number }))
        proxy_protocol = optional(object({ version = optional(string) }))
      }))
    })))

    # Namespaces this destination rule is exported to. Default is all namespaces.
    export_to = optional(list(string))

    # Selects the pods/VMs this rule applies to, by label (istio.type.v1beta1 selector,
    # JSON `matchLabels`). Matched at runtime by istiod; not an OpenMCF foreign key.
    workload_selector = optional(object({
      match_labels = optional(map(string))
    }))
  })
}
