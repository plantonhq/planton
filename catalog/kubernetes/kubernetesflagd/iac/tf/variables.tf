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
  description = "KubernetesFlagd specification"
  type = object({
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    namespace = string

    create_namespace = optional(bool, false)
    image = optional(object({
      repository        = optional(string)
      tag               = optional(string)
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
    sources = list(object({
      config_map = optional(object({
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        config_map_name = string

        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        key = string

        watcher = optional(string, "")
      }))
      http = optional(object({
        url               = string
        auth_header       = optional(string, "")
        headers           = optional(map(string), {})
        sensitive_headers = optional(map(string), {})
        interval_seconds  = optional(number)
        interval_seed     = optional(string, "")
        timeout_seconds   = optional(number)
        oauth = optional(object({
          client_id     = string
          client_secret = string
          token_url     = string
        }))
      }))
      grpc = optional(object({
        target = string
        tls    = optional(bool, false)
        ca_cert_secret = optional(object({
          # The name of the Kubernetes Secret.
          name = optional(string, "")

          # The key within the Kubernetes Secret.
          key = optional(string, "")
        }))
        provider_id         = optional(string, "")
        max_msg_size        = optional(number)
        incremental_updates = optional(bool, false)
        headers             = optional(map(string), {})
        sensitive_headers   = optional(map(string), {})
        selector            = optional(string, "")
      }))
      feature_flag = optional(object({
        namespace = optional(string, "")
        name      = string
      }))
      google_storage = optional(object({
        bucket           = string
        object           = string
        interval_seconds = optional(number)
        interval_seed    = optional(string, "")
      }))
      azure_blob = optional(object({
        bucket           = string
        object           = string
        interval_seconds = optional(number)
        interval_seed    = optional(string, "")
      }))
      s3 = optional(object({
        bucket           = string
        object           = string
        interval_seconds = optional(number)
        interval_seed    = optional(string, "")
      }))
    }))
    server = optional(object({
      port            = optional(number)
      management_port = optional(number)
      sync_port       = optional(number)
      ofrep_port      = optional(number)
      tls_secret_name = optional(string, "")
      service_type    = optional(string)
    }))
    evaluation = optional(object({
      context_values      = optional(map(string), {})
      context_from_header = optional(map(string), {})
      cors_origins        = optional(list(string), [])
    }))
    ofrep_sse = optional(object({
      enabled                  = optional(bool)
      inactivity_delay_seconds = optional(number)
      public_url               = optional(string, "")
    }))
    sync = optional(object({
      http_enabled     = optional(bool)
      disable_metadata = optional(bool, false)
      stream_deadline  = optional(string, "")
    }))
    log = optional(object({
      format = optional(string)
      debug  = optional(bool, false)
    }))
    telemetry = optional(object({
      metrics_exporter   = optional(string, "")
      otel_collector_uri = optional(string, "")
      otel_ca_cert_secret = optional(object({
        # The name of the Kubernetes Secret.
        name = optional(string, "")

        # The key within the Kubernetes Secret.
        key = optional(string, "")
      }))
      otel_client_tls_secret_name = optional(string, "")
      otel_reload_interval        = optional(string, "")
    }))
    limits = optional(object({
      max_request_body_bytes   = optional(number)
      max_request_header_bytes = optional(number)
    }))
    scheduling = optional(object({
      # Simple hard node filter: every listed label must match the node. The right tool
      # for "run on the GPU pool" — reach for node_affinity only when you need operators
      # (In/NotIn/Exists) or soft preferences.
      node_selector = optional(map(string), {})

      # Taint tolerations. A toleration does not attract pods to tainted nodes — it only
      # permits scheduling there; pair with node_selector or affinity to target them.
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

      # Expressive node selection: hard requirements and weighted soft preferences over
      # node labels.
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

      # Attract pods toward nodes/zones already running matching pods (co-location with
      # a cache, for example).
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

      # Repel pods from nodes/zones already running matching pods — the classic
      # high-availability pattern is anti-affinity on the workload's own labels across
      # `kubernetes.io/hostname`.
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

      # Even distribution of replicas across topology domains (zones, hosts). Preferred
      # over hostname anti-affinity for large replica counts because skew is bounded
      # rather than binary.
      topology_spread_constraints = optional(list(object({
        # Maximum allowed difference in matching-pod counts between any two domains.
        # 1 is the strictest even spread.
        max_skew = optional(number, 0)

        # Node label defining the domains to spread across (e.g.
        # "topology.kubernetes.io/zone").
        topology_key = string

        # What happens when the constraint cannot be met: "DoNotSchedule" (hard — pod
        # stays Pending) or "ScheduleAnyway" (soft — scheduler minimizes skew).
        when_unsatisfiable = string

        # Labels selecting the pods counted per domain. Omit to have the module default to
        # the workload's own selector labels — self-spreading, the overwhelmingly common
        # intent.
        match_labels = optional(map(string), {})
      })), [])

      # Hand pods to a non-default scheduler installed in the cluster. Leave empty for
      # the standard scheduler.
      scheduler_name = optional(string, "")
    }))
    service_account = optional(object({
      annotations   = optional(map(string), {})
      existing_name = optional(string, "")
    }))
    pod_annotations = optional(map(string), {})
    pod_labels      = optional(map(string), {})
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
    metrics = optional(object({
      service_monitor_enabled = optional(bool, false)
      service_monitor_labels  = optional(map(string), {})
    }))
  })
}
