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
  description = "KubernetesGoFeatureFlag specification"
  type = object({
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    namespace = string

    create_namespace = optional(bool, false)
    chart_version    = optional(string)
    image = optional(object({
      repository        = optional(string)
      tag               = optional(string, "")
      fips              = optional(bool, false)
      pull_policy       = optional(string)
      pull_secret_names = optional(list(string), [])
    }))
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
    hpa = optional(object({
      enabled                           = optional(bool, false)
      min_replicas                      = optional(number)
      max_replicas                      = optional(number)
      target_cpu_utilization_percent    = optional(number)
      target_memory_utilization_percent = optional(number)
    }))
    pdb = optional(object({
      enabled         = optional(bool, false)
      min_available   = optional(string, "")
      max_unavailable = optional(string, "")
    }))
    server = optional(object({
      port            = optional(number)
      monitoring_port = optional(number)
      service_type    = optional(string)
    }))
    log = optional(object({
      level  = optional(string)
      format = optional(string)
    }))
    authorized_keys = optional(object({
      admin      = optional(list(string), [])
      evaluation = optional(list(string), [])
    }))
    flag_source = optional(object({
      retrievers = list(object({
        config_map = optional(object({
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          config_map_name = string

          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          key = string

          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          namespace = optional(string, "")
        }))
        http = optional(object({
          url               = string
          method            = optional(string, "")
          body              = optional(string, "")
          headers           = optional(map(string), {})
          sensitive_headers = optional(map(string), {})
          timeout_ms        = optional(number)
        }))
        github = optional(object({
          repository_slug = string
          path            = string
          branch          = optional(string, "")
          token           = optional(string, "")
          base_url        = optional(string, "")
          timeout_ms      = optional(number)
        }))
        gitlab = optional(object({
          repository_slug = string
          path            = string
          branch          = optional(string, "")
          token           = optional(string, "")
          base_url        = optional(string, "")
          timeout_ms      = optional(number)
        }))
        bitbucket = optional(object({
          repository_slug = string
          path            = string
          branch          = optional(string, "")
          token           = optional(string, "")
          base_url        = optional(string, "")
          timeout_ms      = optional(number)
        }))
        s3 = optional(object({
          bucket = string
          item   = string
        }))
        google_storage = optional(object({
          bucket = string
          object = string
        }))
        azure_blob_storage = optional(object({
          account_name = string
          account_key  = optional(string, "")
          container    = string
          object       = string
        }))
        mongodb = optional(object({
          uri        = string
          database   = string
          collection = string
        }))
        redis = optional(object({
          options = object({
            addr                    = string
            network                 = optional(string, "")
            username                = optional(string, "")
            password                = optional(string, "")
            db                      = optional(number, 0)
            tls_enabled             = optional(bool, false)
            protocol                = optional(number)
            client_name             = optional(string, "")
            identity_suffix         = optional(string, "")
            disable_identity        = optional(bool, false)
            max_retries             = optional(number)
            min_retry_backoff_ms    = optional(number)
            max_retry_backoff_ms    = optional(number)
            dial_timeout_ms         = optional(number)
            read_timeout_ms         = optional(number)
            write_timeout_ms        = optional(number)
            context_timeout_enabled = optional(bool, false)
            pool_fifo               = optional(bool, false)
            pool_size               = optional(number)
            pool_timeout_ms         = optional(number)
            min_idle_conns          = optional(number, 0)
            max_idle_conns          = optional(number, 0)
            conn_max_idle_time_ms   = optional(number)
            conn_max_lifetime_ms    = optional(number)
          })
          prefix = optional(string, "")
        }))
        postgresql = optional(object({
          uri     = string
          table   = string
          columns = optional(map(string), {})
        }))
      }))
      notifiers = optional(list(object({
        slack = optional(object({
          webhook_url = string
        }))
        microsoft_teams = optional(object({
          webhook_url = string
        }))
        discord = optional(object({
          webhook_url = string
        }))
        webhook = optional(object({
          endpoint_url      = string
          secret            = optional(string, "")
          meta              = optional(map(string), {})
          headers           = optional(map(string), {})
          sensitive_headers = optional(map(string), {})
        }))
      })), [])
      exporters                     = optional(any, [])
      file_format                   = optional(string, "")
      polling_interval_ms           = optional(number)
      start_with_retriever_error    = optional(bool)
      enable_polling_jitter         = optional(bool, false)
      disable_notifier_on_init      = optional(bool, false)
      evaluation_context_enrichment = optional(any)
    }))
    flag_sets = optional(object({
      items = any
    }))
    ofrep_event_stream = optional(object({
      base_url             = optional(string, "")
      inactivity_delay_sec = optional(number, 0)
    }))
    telemetry = optional(object({
      otlp_endpoint       = optional(string, "")
      otlp_protocol       = optional(string, "")
      sdk_disabled        = optional(bool, false)
      service_name        = optional(string, "")
      traces_sampler      = optional(string, "")
      resource_attributes = optional(map(string), {})
      jaeger_sampler = optional(object({
        manager_host_port = optional(string, "")
        refresh_interval  = optional(string, "")
        max_operations    = optional(number, 0)
      }))
      traces_sampler_arg = optional(string, "")
    }))
    swagger = optional(object({
      enabled = optional(bool, false)
      host    = optional(string, "")
    }))
    runtime = optional(object({
      hide_banner                    = optional(bool, false)
      enable_pprof                   = optional(bool, false)
      disable_version_header         = optional(bool, false)
      enable_bulk_metric_flag_names  = optional(bool, false)
      disable_flag_details_in_stream = optional(bool, false)
      exporter_clean_queue_interval  = optional(string, "")
      env_variable_prefix            = optional(string)
    }))
    metrics = optional(object({
      service_monitor_enabled = optional(bool, false)
      service_monitor_labels  = optional(map(string), {})
    }))
    scheduling = optional(object({
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
      node_affinity = optional(object({
        # Hard requirement. The outer list ORs its terms; expressions within one term AND.
        required = optional(list(object({
          match_expressions = list(object({
            # Node label key, e.g. "topology.kubernetes.io/zone".
            key = string

            # Operator: "In"/"NotIn" (value set), "Exists"/"DoesNotExist" (key presence), or
            # "Gt"/"Lt" (single integer value, as strings — the Kubernetes API convention).
            operator = string

            # Values for the operator: required non-empty for In/NotIn, exactly one integer
            # string for Gt/Lt, and must be empty for Exists/DoesNotExist.
            values = optional(list(string), [])
          }))
        })), [])

        # Weighted soft preferences.
        preferred = optional(list(object({
          # Preference weight, 1–100. Higher weights dominate placement scoring.
          weight = optional(number, 0)

          term = object({
            match_expressions = list(object({
              # Node label key, e.g. "topology.kubernetes.io/zone".
              key = string

              # Operator: "In"/"NotIn" (value set), "Exists"/"DoesNotExist" (key presence), or
              # "Gt"/"Lt" (single integer value, as strings — the Kubernetes API convention).
              operator = string

              # Values for the operator: required non-empty for In/NotIn, exactly one integer
              # string for Gt/Lt, and must be empty for Exists/DoesNotExist.
              values = optional(list(string), [])
            }))
          })
        })), [])
      }))
      pod_affinity = optional(object({
        # Hard rules — unschedulable until satisfied. Use sparingly; they can deadlock rollouts.
        required = optional(list(object({
          # Labels of the pods to match against — for self-anti-affinity, the workload's own
          # selector labels (exported as the `selector_labels` output).
          match_labels = optional(map(string), {})

          # Node label defining the domain: "kubernetes.io/hostname" separates by node,
          # "topology.kubernetes.io/zone" by zone.
          topology_key = string

          # Namespaces whose pods are considered. Empty means the workload's own namespace.
          namespaces = optional(list(string), [])
        })), [])

        # Weighted soft rules — the scheduler's tiebreakers.
        preferred = optional(list(object({
          # Preference weight, 1–100.
          weight = optional(number, 0)

          term = object({
            # Labels of the pods to match against — for self-anti-affinity, the workload's own
            # selector labels (exported as the `selector_labels` output).
            match_labels = optional(map(string), {})

            # Node label defining the domain: "kubernetes.io/hostname" separates by node,
            # "topology.kubernetes.io/zone" by zone.
            topology_key = string

            # Namespaces whose pods are considered. Empty means the workload's own namespace.
            namespaces = optional(list(string), [])
          })
        })), [])
      }))
      pod_anti_affinity = optional(object({
        # Hard rules — unschedulable until satisfied. Use sparingly; they can deadlock rollouts.
        required = optional(list(object({
          # Labels of the pods to match against — for self-anti-affinity, the workload's own
          # selector labels (exported as the `selector_labels` output).
          match_labels = optional(map(string), {})

          # Node label defining the domain: "kubernetes.io/hostname" separates by node,
          # "topology.kubernetes.io/zone" by zone.
          topology_key = string

          # Namespaces whose pods are considered. Empty means the workload's own namespace.
          namespaces = optional(list(string), [])
        })), [])

        # Weighted soft rules — the scheduler's tiebreakers.
        preferred = optional(list(object({
          # Preference weight, 1–100.
          weight = optional(number, 0)

          term = object({
            # Labels of the pods to match against — for self-anti-affinity, the workload's own
            # selector labels (exported as the `selector_labels` output).
            match_labels = optional(map(string), {})

            # Node label defining the domain: "kubernetes.io/hostname" separates by node,
            # "topology.kubernetes.io/zone" by zone.
            topology_key = string

            # Namespaces whose pods are considered. Empty means the workload's own namespace.
            namespaces = optional(list(string), [])
          })
        })), [])
      }))
    }))
    service_account = optional(object({
      annotations   = optional(map(string), {})
      existing_name = optional(string, "")
    }))
    pod_annotations = optional(map(string), {})
    pod_labels      = optional(map(string), {})
    common_labels   = optional(map(string), {})
    pod_security_context = optional(object({
      # UID all container processes run as unless overridden per container.
      run_as_user = optional(number)

      # Primary GID all container processes run as unless overridden per container.
      run_as_group = optional(number)

      # Refuse to start any container whose effective user is root.
      run_as_non_root = optional(bool)

      # GID that owns mounted volumes and is added to every container's supplemental
      # groups — the standard fix for "permission denied" on persistent volumes written
      # by non-root apps.
      fs_group = optional(number)

      # When volume ownership is re-chowned to fs_group: "Always" (default) or
      # "OnRootMismatch" (skip the recursive chown when the root already matches —
      # dramatically faster pod starts on large volumes).
      fs_group_change_policy = optional(string, "")

      # Additional group IDs applied to all container processes.
      supplemental_groups = optional(list(number), [])

      # Kernel parameters set for the pod. Only safe sysctls (or those the cluster
      # administrator has allow-listed on the kubelet) are admitted.
      sysctls = optional(list(object({
        # Sysctl name, e.g. "net.core.somaxconn".
        name = string

        # Sysctl value, e.g. "1024".
        value = string
      })), [])

      # Pod-wide seccomp profile; containers may override with their own.
      seccomp_profile = optional(object({
        # Profile type: "RuntimeDefault" (the container runtime's default filter — the
        # recommended baseline), "Unconfined" (no filtering), or "Localhost" (a profile
        # file installed on the node, named via localhost_profile).
        type = string

        # Path of the profile file relative to the node's seccomp profile root. Required
        # when (and only meaningful when) type is "Localhost".
        localhost_profile = optional(string, "")
      }))
    }))
    container_security_context = optional(object({
      # Runs the container with full host access — equivalent to root on the node.
      # Required by some node-level agents (device managers, network plugins). Never
      # combine with untrusted images.
      privileged = optional(bool, false)

      # UID the container process runs as. Overrides the image's USER directive.
      run_as_user = optional(number)

      # Primary GID the container process runs as.
      run_as_group = optional(number)

      # Refuses to start the container if its effective user is root. The standard
      # baseline hardening — it catches images that silently default to UID 0.
      run_as_non_root = optional(bool)

      # Mounts the container's root filesystem read-only. Pair with EmptyDir mounts for
      # paths the app must write (e.g. /tmp).
      read_only_root_filesystem = optional(bool)

      # Whether the process can gain more privileges than its parent (setuid binaries,
      # file capabilities). The restricted Pod Security Standard requires this to be
      # false. Always true when `privileged` is set, so leave it unset in that case.
      allow_privilege_escalation = optional(bool)

      # Linux capabilities to add or drop. The restricted profile drops ALL and adds
      # back only NET_BIND_SERVICE when needed. Capability names are uppercase without
      # the CAP_ prefix (e.g. "NET_ADMIN", "SYS_TIME").
      capabilities = optional(object({
        # Capabilities to add (e.g. "NET_BIND_SERVICE").
        add = optional(list(string), [])

        # Capabilities to drop. Use ["ALL"] as the hardened baseline.
        drop = optional(list(string), [])
      }))

      # Seccomp syscall filter for the container. "RuntimeDefault" is the hardened
      # baseline; "Localhost" selects a node-local profile file via `localhost_profile`.
      seccomp_profile = optional(object({
        # Profile type: "RuntimeDefault" (the container runtime's default filter — the
        # recommended baseline), "Unconfined" (no filtering), or "Localhost" (a profile
        # file installed on the node, named via localhost_profile).
        type = string

        # Path of the profile file relative to the node's seccomp profile root. Required
        # when (and only meaningful when) type is "Localhost".
        localhost_profile = optional(string, "")
      }))
    }))
    extra_env = optional(map(string), {})
    extra_env_from_secret = optional(map(object({
      # The name of the Kubernetes Secret.
      name = optional(string, "")

      # The key within the Kubernetes Secret.
      key = optional(string, "")
    })), {})
    helm_values = optional(string, "")
  })
}
