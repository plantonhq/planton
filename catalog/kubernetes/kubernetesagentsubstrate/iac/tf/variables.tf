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
  description = "KubernetesAgentSubstrate specification"
  type = object({
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    namespace = string

    create_namespace = optional(bool, false)
    crds = optional(object({
      install           = optional(bool)
      keep_on_uninstall = optional(bool)
    }))
    platform = object({
      gke = optional(object({
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        cluster_id = string

        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        api_server_service_account = optional(string, "")

        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        node_agent_service_account = optional(string, "")

        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        snapshot_transfer_service_account = optional(string, "")
      }))
      generic = optional(object({
        service_account_issuer = string
      }))
    })
    database = object({
      bundled = optional(object({
        storage_class = optional(string, "")
        volume_size   = optional(string)
        resources = optional(object({
          # The resource limits for the container.
          # Specify the maximum amount of CPU and memory that the container can use.
          limits = optional(object({
            cpu    = optional(string, "")
            memory = optional(string, "")
          }))

          # The resource requests for the container.
          # Specify the minimum amount of CPU and memory that the container is guaranteed.
          requests = optional(object({
            cpu    = optional(string, "")
            memory = optional(string, "")
          }))
        }))
      }))
      external = optional(object({
        read_write_connection_string = object({
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          name = string

          key = string
        })
        owner_connection_string = optional(object({
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          name = string

          key = string
        }))
        roles = optional(object({
          read_write = optional(string)
          owner      = optional(string)
          schema     = optional(string)
        }))
        server_ca_certificate = optional(string, "")
      }))
      cloud_sql = optional(object({
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        instance_connection_name = string

        iam_authentication = optional(bool)
        ip_type            = optional(string, "")
        read_write_connection_string = optional(object({
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          name = string

          key = string
        }))
        owner_connection_string = optional(object({
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          name = string

          key = string
        }))
        roles = optional(object({
          read_write = optional(string)
          owner      = optional(string)
          schema     = optional(string)
        }))
      }))
      pool_max_connections = optional(number)
    })
    storage = object({
      gcs = optional(object({
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        bucket = string
      }))
      s3 = optional(object({
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        endpoint = string

        bucket           = string
        region           = optional(string)
        force_path_style = optional(bool)
        credentials = object({
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          secret_name = string

          access_key_id_key     = optional(string)
          secret_access_key_key = optional(string)
        })
      }))
      snapshot_prefix = optional(string, "")
    })
    api_server = optional(object({
      replicas = optional(number)
      resources = optional(object({
        # The resource limits for the container.
        # Specify the maximum amount of CPU and memory that the container can use.
        limits = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))

        # The resource requests for the container.
        # Specify the minimum amount of CPU and memory that the container is guaranteed.
        requests = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
      }))
      log_level = optional(string, "")
      authorization = optional(object({
        bootstrap_owners = optional(list(string), [])
      }))
      actor_jwt = optional(object({
        algorithm = optional(string, "")
        issuer    = optional(string, "")
      }))
      drain_delay              = optional(string)
      drain_timeout            = optional(string)
      template_resync_interval = optional(string)
    }))
    controller = optional(object({
      resources = optional(object({
        # The resource limits for the container.
        # Specify the maximum amount of CPU and memory that the container can use.
        limits = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))

        # The resource requests for the container.
        # Specify the minimum amount of CPU and memory that the container is guaranteed.
        requests = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
      }))
      log_level               = optional(string, "")
      worker_sync_concurrency = optional(number)
    }))
    node_agent = optional(object({
      resources = optional(object({
        # The resource limits for the container.
        # Specify the maximum amount of CPU and memory that the container can use.
        limits = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))

        # The resource requests for the container.
        # Specify the minimum amount of CPU and memory that the container is guaranteed.
        requests = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
      }))
      log_level     = optional(string, "")
      node_selector = optional(map(string), {})
      tolerations = optional(list(object({
        # Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.
        key = optional(string, "")

        # How key/value match: "Equal" (default — value must match too) or "Exists"
        # (key presence alone matches).
        operator = optional(string, "")

        # Taint value to match when operator is "Equal".
        value = optional(string, "")

        # Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
        # "NoExecute". Empty tolerates all effects for the key.
        effect = optional(string, "")

        # For "NoExecute" taints only: how many seconds already-running pods stay bound
        # after the taint appears. Unset means tolerate forever.
        toleration_seconds = optional(number)
      })), [])
      actor_stats_poll_interval = optional(string)
      drain_delay               = optional(string)
      drain_timeout             = optional(string)
      image_cache = optional(object({
        gc_period    = optional(string)
        high_percent = optional(number)
        low_percent  = optional(number)
        max_bytes    = optional(number)
        min_age      = optional(string)
        dry_run      = optional(bool, false)
      }))
      asset_staging_capacity_gib = optional(number)
      snapshots = optional(object({
        upload_delay          = optional(string)
        detach_delay          = optional(string)
        upload_timeout        = optional(string)
        upload_throughput_mib = optional(number)
        upload_workers        = optional(number)
        detach_workers        = optional(number)
        report_workers        = optional(number)
        report_give_up_after  = optional(string)
        upload_memory_budget  = optional(string)
      }))
      sandbox_asset_cache = optional(object({
        min_age       = optional(string)
        fetch_timeout = optional(string)
        eviction = optional(object({
          gc_period = optional(string)
          max_bytes = optional(number)
          dry_run   = optional(bool, false)
        }))
      }))
      snapshot_cache = optional(object({
        disabled      = optional(bool, false)
        min_age       = optional(string)
        fetch_timeout = optional(string)
        eviction = optional(object({
          gc_period = optional(string)
          max_bytes = optional(number)
          dry_run   = optional(bool, false)
        }))
      }))
    }))
    router = optional(object({
      dataplane = optional(string, "")
      replicas  = optional(number)
      resources = optional(object({
        # The resource limits for the container.
        # Specify the maximum amount of CPU and memory that the container can use.
        limits = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))

        # The resource requests for the container.
        # Specify the minimum amount of CPU and memory that the container is guaranteed.
        requests = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
      }))
      log_level                 = optional(string, "")
      allowed_client_spiffe_ids = optional(list(string), [])
      route_timeout             = optional(string)
      parked_requests = optional(object({
        budget         = optional(string)
        max            = optional(number)
        retry_interval = optional(string)
        retry_factor   = optional(number)
        retry_jitter   = optional(number)
      }))
      ext_proc_max_requests = optional(number)
      drain_delay           = optional(string)
      drain_timeout         = optional(string, "")
    }))
    egress = object({
      credential_provider = object({
        disabled = optional(object({}))
        kubernetes_secrets = optional(object({
          grants = optional(list(object({
            atespace           = string
            allowed_namespaces = list(string)
          })), [])
          resources = optional(object({
            # The resource limits for the container.
            # Specify the maximum amount of CPU and memory that the container can use.
            limits = optional(object({
              cpu    = optional(string, "")
              memory = optional(string, "")
            }))

            # The resource requests for the container.
            # Specify the minimum amount of CPU and memory that the container is guaranteed.
            requests = optional(object({
              cpu    = optional(string, "")
              memory = optional(string, "")
            }))
          }))
          log_level      = optional(string, "")
          kube_api_qps   = optional(number)
          kube_api_burst = optional(number)
          drain_grace    = optional(string)
        }))
        external = optional(object({
          name    = string
          address = string
        }))
      })
      replicas = optional(number)
      resources = optional(object({
        # The resource limits for the container.
        # Specify the maximum amount of CPU and memory that the container can use.
        limits = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))

        # The resource requests for the container.
        # Specify the minimum amount of CPU and memory that the container is guaranteed.
        requests = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
      }))
      log_level                 = optional(string, "")
      policy_cache_ttl          = optional(string)
      intercept_certificate_ttl = optional(string)
      envoy_concurrency         = optional(number)
      additional_ext_proc_service = optional(object({
        namespace = string
        service   = string
        port      = optional(number, 0)
      }))
      drain_delay = optional(string)
    })
    pod_certificates = optional(object({
      resources = optional(object({
        # The resource limits for the container.
        # Specify the maximum amount of CPU and memory that the container can use.
        limits = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))

        # The resource requests for the container.
        # Specify the minimum amount of CPU and memory that the container is guaranteed.
        requests = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
      }))
      workers_per_signer = optional(number)
      kube_api_qps       = optional(number)
      kube_api_burst     = optional(number)
    }))
    telemetry = optional(object({
      otlp_endpoint             = optional(string, "")
      traces_exporters          = optional(list(string), [])
      metrics_exporters         = optional(list(string), [])
      logs_exporters            = optional(list(string), [])
      traces_sampler            = optional(string, "")
      traces_sampler_ratio      = optional(number)
      metric_export_interval_ms = optional(number)
      metric_export_timeout_ms  = optional(number)
      worker_traces_exporters   = optional(list(string), [])
      worker_metrics_exporters  = optional(list(string), [])
      worker_logs_exporters     = optional(list(string), [])
    }))
    cluster_size         = optional(string, "")
    cordon_control_plane = optional(bool, false)
    image_registry       = optional(string, "")
    image_pull_secrets   = optional(list(string), [])
    version              = optional(string)
    preview_gates        = optional(list(string), [])
    identity_server = optional(object({
      replicas = optional(number)
      resources = optional(object({
        # The resource limits for the container.
        # Specify the maximum amount of CPU and memory that the container can use.
        limits = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))

        # The resource requests for the container.
        # Specify the minimum amount of CPU and memory that the container is guaranteed.
        requests = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
      }))
      log_level   = optional(string, "")
      drain_grace = optional(string)
    }))
  })
}
