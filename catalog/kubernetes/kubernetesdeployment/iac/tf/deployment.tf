# The apps/v1 Deployment. Every container (app and sidecars) renders through
# the same dynamic block from locals.all_containers, so semantics are identical
# for all container roles — mirroring the Pulumi module's shared builders.
#
# Identity and exposure are composed, not created: pods run as the
# ServiceAccount named in spec.pod.service_account, and external exposure
# attaches through first-class ingress kinds referencing this workload's
# Service. This module never creates ServiceAccounts, RBAC objects,
# certificates, gateways, or routes.

resource "kubernetes_deployment_v1" "this" {
  metadata {
    name      = var.metadata.name
    namespace = local.namespace
    labels    = local.final_labels
  }

  # The apply waits for the rollout it starts, and fails when the new pods
  # never become ready. An apply that changes nothing starts no rollout and
  # does not wait, so it cannot notice a rollout an earlier apply left failing.
  wait_for_rollout = true

  spec {
    replicas = local.replicas

    # The selector is the workload's immutable pod-selection identity — it must
    # never include mutable labels like version.
    selector {
      match_labels = local.selector_labels
    }

    dynamic "strategy" {
      for_each = try(var.spec.availability.strategy, null) != null ? [var.spec.availability.strategy] : []
      content {
        type = local.strategy_type

        # A Recreate strategy must NOT carry a rolling_update block — the API
        # server rejects the combination.
        dynamic "rolling_update" {
          for_each = (
            local.strategy_type == "RollingUpdate" &&
            (try(strategy.value.max_unavailable, "") != "" || try(strategy.value.max_surge, "") != "")
          ) ? [strategy.value] : []
          content {
            max_unavailable = try(rolling_update.value.max_unavailable, "") != "" ? rolling_update.value.max_unavailable : null
            max_surge       = try(rolling_update.value.max_surge, "") != "" ? rolling_update.value.max_surge : null
          }
        }
      }
    }

    min_ready_seconds         = try(var.spec.availability.min_ready_seconds, null)
    revision_history_limit    = try(var.spec.availability.revision_history_limit, null)
    progress_deadline_seconds = try(var.spec.availability.progress_deadline_seconds, null)
    paused                    = try(var.spec.availability.paused, false)

    template {
      metadata {
        labels      = local.pod_template_labels
        annotations = length(try(var.spec.pod.annotations, {})) > 0 ? var.spec.pod.annotations : null
      }

      spec {
        service_account_name            = try(var.spec.pod.service_account, "") != "" ? var.spec.pod.service_account : null
        automount_service_account_token = try(var.spec.pod.automount_service_account_token, null)
        termination_grace_period_seconds = try(var.spec.pod.termination_grace_period_seconds, null)
        dns_policy                      = try(var.spec.pod.dns_policy, "") != "" ? var.spec.pod.dns_policy : null
        host_network                    = try(var.spec.pod.host_network, false)
        host_pid                        = try(var.spec.pod.host_pid, false)
        priority_class_name             = try(var.spec.pod.priority_class_name, "") != "" ? var.spec.pod.priority_class_name : null
        runtime_class_name              = try(var.spec.pod.runtime_class_name, "") != "" ? var.spec.pod.runtime_class_name : null
        node_selector                   = length(try(var.spec.pod.scheduling.node_selector, {})) > 0 ? var.spec.pod.scheduling.node_selector : null
        scheduler_name                  = try(var.spec.pod.scheduling.scheduler_name, "") != "" ? var.spec.pod.scheduling.scheduler_name : null

        dynamic "image_pull_secrets" {
          for_each = local.image_pull_secret_names
          content {
            name = image_pull_secrets.value
          }
        }

        dynamic "dns_config" {
          for_each = try(var.spec.pod.dns_config, null) != null ? [var.spec.pod.dns_config] : []
          content {
            nameservers = try(dns_config.value.nameservers, [])
            searches    = try(dns_config.value.searches, [])
            dynamic "option" {
              for_each = try(dns_config.value.options, [])
              content {
                name  = option.value.name
                value = try(option.value.value, "") != "" ? option.value.value : null
              }
            }
          }
        }

        dynamic "host_aliases" {
          for_each = try(var.spec.pod.host_aliases, [])
          content {
            ip        = host_aliases.value.ip
            hostnames = host_aliases.value.hostnames
          }
        }

        dynamic "toleration" {
          for_each = try(var.spec.pod.scheduling.tolerations, [])
          content {
            key                = try(toleration.value.key, "") != "" ? toleration.value.key : null
            operator           = try(toleration.value.operator, "") != "" ? toleration.value.operator : null
            value              = try(toleration.value.value, "") != "" ? toleration.value.value : null
            effect             = try(toleration.value.effect, "") != "" ? toleration.value.effect : null
            toleration_seconds = try(toleration.value.toleration_seconds, null)
          }
        }

        dynamic "affinity" {
          for_each = (
            try(var.spec.pod.scheduling.node_affinity, null) != null ||
            try(var.spec.pod.scheduling.pod_affinity, null) != null ||
            try(var.spec.pod.scheduling.pod_anti_affinity, null) != null
          ) ? [var.spec.pod.scheduling] : []
          content {
            dynamic "node_affinity" {
              for_each = try(affinity.value.node_affinity, null) != null ? [affinity.value.node_affinity] : []
              content {
                dynamic "required_during_scheduling_ignored_during_execution" {
                  for_each = length(try(node_affinity.value.required, [])) > 0 ? [node_affinity.value.required] : []
                  content {
                    dynamic "node_selector_term" {
                      for_each = required_during_scheduling_ignored_during_execution.value
                      content {
                        dynamic "match_expressions" {
                          for_each = node_selector_term.value.match_expressions
                          content {
                            key      = match_expressions.value.key
                            operator = match_expressions.value.operator
                            values   = length(try(match_expressions.value.values, [])) > 0 ? match_expressions.value.values : null
                          }
                        }
                      }
                    }
                  }
                }
                dynamic "preferred_during_scheduling_ignored_during_execution" {
                  for_each = try(node_affinity.value.preferred, [])
                  content {
                    weight = preferred_during_scheduling_ignored_during_execution.value.weight
                    preference {
                      dynamic "match_expressions" {
                        for_each = preferred_during_scheduling_ignored_during_execution.value.term.match_expressions
                        content {
                          key      = match_expressions.value.key
                          operator = match_expressions.value.operator
                          values   = length(try(match_expressions.value.values, [])) > 0 ? match_expressions.value.values : null
                        }
                      }
                    }
                  }
                }
              }
            }

            dynamic "pod_affinity" {
              for_each = try(affinity.value.pod_affinity, null) != null ? [affinity.value.pod_affinity] : []
              content {
                dynamic "required_during_scheduling_ignored_during_execution" {
                  for_each = try(pod_affinity.value.required, [])
                  content {
                    topology_key = required_during_scheduling_ignored_during_execution.value.topology_key
                    namespaces   = length(try(required_during_scheduling_ignored_during_execution.value.namespaces, [])) > 0 ? required_during_scheduling_ignored_during_execution.value.namespaces : null
                    label_selector {
                      match_labels = required_during_scheduling_ignored_during_execution.value.match_labels
                    }
                  }
                }
                dynamic "preferred_during_scheduling_ignored_during_execution" {
                  for_each = try(pod_affinity.value.preferred, [])
                  content {
                    weight = preferred_during_scheduling_ignored_during_execution.value.weight
                    pod_affinity_term {
                      topology_key = preferred_during_scheduling_ignored_during_execution.value.term.topology_key
                      namespaces   = length(try(preferred_during_scheduling_ignored_during_execution.value.term.namespaces, [])) > 0 ? preferred_during_scheduling_ignored_during_execution.value.term.namespaces : null
                      label_selector {
                        match_labels = preferred_during_scheduling_ignored_during_execution.value.term.match_labels
                      }
                    }
                  }
                }
              }
            }

            dynamic "pod_anti_affinity" {
              for_each = try(affinity.value.pod_anti_affinity, null) != null ? [affinity.value.pod_anti_affinity] : []
              content {
                dynamic "required_during_scheduling_ignored_during_execution" {
                  for_each = try(pod_anti_affinity.value.required, [])
                  content {
                    topology_key = required_during_scheduling_ignored_during_execution.value.topology_key
                    namespaces   = length(try(required_during_scheduling_ignored_during_execution.value.namespaces, [])) > 0 ? required_during_scheduling_ignored_during_execution.value.namespaces : null
                    label_selector {
                      match_labels = required_during_scheduling_ignored_during_execution.value.match_labels
                    }
                  }
                }
                dynamic "preferred_during_scheduling_ignored_during_execution" {
                  for_each = try(pod_anti_affinity.value.preferred, [])
                  content {
                    weight = preferred_during_scheduling_ignored_during_execution.value.weight
                    pod_affinity_term {
                      topology_key = preferred_during_scheduling_ignored_during_execution.value.term.topology_key
                      namespaces   = length(try(preferred_during_scheduling_ignored_during_execution.value.term.namespaces, [])) > 0 ? preferred_during_scheduling_ignored_during_execution.value.term.namespaces : null
                      label_selector {
                        match_labels = preferred_during_scheduling_ignored_during_execution.value.term.match_labels
                      }
                    }
                  }
                }
              }
            }
          }
        }

        dynamic "topology_spread_constraint" {
          for_each = try(var.spec.pod.scheduling.topology_spread_constraints, [])
          content {
            max_skew           = topology_spread_constraint.value.max_skew
            topology_key       = topology_spread_constraint.value.topology_key
            when_unsatisfiable = topology_spread_constraint.value.when_unsatisfiable
            # Empty match_labels self-spreads on the workload's own selector —
            # the overwhelmingly common intent, mirrored in the Pulumi module.
            label_selector {
              match_labels = length(try(topology_spread_constraint.value.match_labels, {})) > 0 ? topology_spread_constraint.value.match_labels : local.selector_labels
            }
          }
        }

        dynamic "security_context" {
          for_each = try(var.spec.pod.security_context, null) != null ? [var.spec.pod.security_context] : []
          content {
            run_as_user            = try(security_context.value.run_as_user, null)
            run_as_group           = try(security_context.value.run_as_group, null)
            run_as_non_root        = try(security_context.value.run_as_non_root, null)
            fs_group               = try(security_context.value.fs_group, null)
            fs_group_change_policy = try(security_context.value.fs_group_change_policy, "") != "" ? security_context.value.fs_group_change_policy : null
            supplemental_groups    = length(try(security_context.value.supplemental_groups, [])) > 0 ? security_context.value.supplemental_groups : null

            dynamic "sysctl" {
              for_each = try(security_context.value.sysctls, [])
              content {
                name  = sysctl.value.name
                value = sysctl.value.value
              }
            }

            dynamic "seccomp_profile" {
              for_each = try(security_context.value.seccomp_profile, null) != null ? [security_context.value.seccomp_profile] : []
              content {
                type              = seccomp_profile.value.type
                localhost_profile = try(seccomp_profile.value.localhost_profile, "") != "" ? seccomp_profile.value.localhost_profile : null
              }
            }
          }
        }

        dynamic "init_container" {
          for_each = local.init_containers
          content {
            name              = init_container.value.name
            image             = "${init_container.value.image.repo}:${init_container.value.image.tag}"
            image_pull_policy = try(init_container.value.image_pull_policy, "") != "" ? init_container.value.image_pull_policy : null
            command           = length(try(init_container.value.command, [])) > 0 ? init_container.value.command : null
            args              = length(try(init_container.value.args, [])) > 0 ? init_container.value.args : null
            working_dir       = try(init_container.value.working_dir, "") != "" ? init_container.value.working_dir : null

            dynamic "env" {
              for_each = try(init_container.value.env.variables, [])
              content {
                name  = env.value.name
                value = try(env.value.value, "") != "" ? env.value.value : null
                dynamic "value_from" {
                  for_each = try(env.value.config_map_key_ref, null) != null ? [env.value.config_map_key_ref] : []
                  content {
                    config_map_key_ref {
                      name     = value_from.value.name
                      key      = value_from.value.key
                      optional = try(value_from.value.optional, false)
                    }
                  }
                }
              }
            }

            dynamic "env" {
              for_each = try(init_container.value.env.secrets, [])
              content {
                name = env.value.name
                value_from {
                  secret_key_ref {
                    name = try(env.value.secret_ref, null) != null ? env.value.secret_ref.name : local.env_secret_name
                    key  = try(env.value.secret_ref, null) != null ? env.value.secret_ref.key : env.value.name
                  }
                }
              }
            }

            dynamic "resources" {
              for_each = try(init_container.value.resources, null) != null ? [init_container.value.resources] : []
              content {
                limits = {
                  for k, v in {
                    cpu    = try(resources.value.limits.cpu, "")
                    memory = try(resources.value.limits.memory, "")
                  } : k => v if v != ""
                }
                requests = {
                  for k, v in {
                    cpu    = try(resources.value.requests.cpu, "")
                    memory = try(resources.value.requests.memory, "")
                  } : k => v if v != ""
                }
              }
            }

            dynamic "volume_mount" {
              for_each = try(init_container.value.volume_mounts, [])
              content {
                name       = volume_mount.value.name
                mount_path = volume_mount.value.mount_path
                read_only  = try(volume_mount.value.read_only, false)
                sub_path   = try(volume_mount.value.sub_path, "") != "" ? volume_mount.value.sub_path : null
              }
            }

            dynamic "security_context" {
              for_each = try(init_container.value.security_context, null) != null ? [init_container.value.security_context] : []
              content {
                privileged                 = try(security_context.value.privileged, false)
                run_as_user                = try(security_context.value.run_as_user, null)
                run_as_group               = try(security_context.value.run_as_group, null)
                run_as_non_root            = try(security_context.value.run_as_non_root, null)
                read_only_root_filesystem  = try(security_context.value.read_only_root_filesystem, null)
                allow_privilege_escalation = try(security_context.value.allow_privilege_escalation, null)
                dynamic "capabilities" {
                  for_each = try(security_context.value.capabilities, null) != null ? [security_context.value.capabilities] : []
                  content {
                    add  = length(try(capabilities.value.add, [])) > 0 ? capabilities.value.add : null
                    drop = length(try(capabilities.value.drop, [])) > 0 ? capabilities.value.drop : null
                  }
                }
                dynamic "seccomp_profile" {
                  for_each = try(security_context.value.seccomp_profile, null) != null ? [security_context.value.seccomp_profile] : []
                  content {
                    type              = seccomp_profile.value.type
                    localhost_profile = try(seccomp_profile.value.localhost_profile, "") != "" ? seccomp_profile.value.localhost_profile : null
                  }
                }
              }
            }
          }
        }

        dynamic "container" {
          for_each = local.all_containers
          content {
            name              = container.value.name
            image             = "${container.value.image.repo}:${container.value.image.tag}"
            image_pull_policy = try(container.value.image_pull_policy, "") != "" ? container.value.image_pull_policy : null
            command           = length(try(container.value.command, [])) > 0 ? container.value.command : null
            args              = length(try(container.value.args, [])) > 0 ? container.value.args : null
            working_dir       = try(container.value.working_dir, "") != "" ? container.value.working_dir : null

            dynamic "port" {
              for_each = try(container.value.ports, [])
              content {
                name           = port.value.name
                container_port = port.value.container_port
                protocol       = try(port.value.network_protocol, "") != "" ? port.value.network_protocol : null
                host_port      = try(port.value.host_port, 0) > 0 ? port.value.host_port : null
              }
            }

            dynamic "env" {
              for_each = try(container.value.env.variables, [])
              content {
                name  = env.value.name
                value = try(env.value.value, "") != "" ? env.value.value : null

                dynamic "value_from" {
                  for_each = try(env.value.config_map_key_ref, null) != null ? [env.value.config_map_key_ref] : []
                  content {
                    config_map_key_ref {
                      name     = value_from.value.name
                      key      = value_from.value.key
                      optional = try(value_from.value.optional, false)
                    }
                  }
                }

                dynamic "value_from" {
                  for_each = try(env.value.field_ref, null) != null ? [env.value.field_ref] : []
                  content {
                    field_ref {
                      api_version = try(value_from.value.api_version, "") != "" ? value_from.value.api_version : null
                      field_path  = value_from.value.field_path
                    }
                  }
                }

                dynamic "value_from" {
                  for_each = try(env.value.resource_field_ref, null) != null ? [env.value.resource_field_ref] : []
                  content {
                    resource_field_ref {
                      container_name = try(value_from.value.container_name, "") != "" ? value_from.value.container_name : null
                      resource       = value_from.value.resource
                      divisor        = try(value_from.value.divisor, "") != "" ? value_from.value.divisor : null
                    }
                  }
                }
              }
            }

            # Literal secret values reference the module-created env Secret by
            # the env var's own name; secretRef entries reference the existing
            # Secret directly.
            dynamic "env" {
              for_each = try(container.value.env.secrets, [])
              content {
                name = env.value.name
                value_from {
                  secret_key_ref {
                    name = try(env.value.secret_ref, null) != null ? env.value.secret_ref.name : local.env_secret_name
                    key  = try(env.value.secret_ref, null) != null ? env.value.secret_ref.key : env.value.name
                  }
                }
              }
            }

            dynamic "env_from" {
              for_each = try(container.value.env.env_from, [])
              content {
                prefix = try(env_from.value.prefix, "") != "" ? env_from.value.prefix : null
                dynamic "config_map_ref" {
                  for_each = try(env_from.value.config_map_ref, null) != null ? [env_from.value.config_map_ref] : []
                  content {
                    name     = config_map_ref.value.name
                    optional = try(config_map_ref.value.optional, false)
                  }
                }
                dynamic "secret_ref" {
                  for_each = try(env_from.value.secret_ref, null) != null ? [env_from.value.secret_ref] : []
                  content {
                    name     = secret_ref.value.name
                    optional = try(secret_ref.value.optional, false)
                  }
                }
              }
            }

            dynamic "resources" {
              for_each = try(container.value.resources, null) != null ? [container.value.resources] : []
              content {
                limits = {
                  for k, v in {
                    cpu    = try(resources.value.limits.cpu, "")
                    memory = try(resources.value.limits.memory, "")
                  } : k => v if v != ""
                }
                requests = {
                  for k, v in {
                    cpu    = try(resources.value.requests.cpu, "")
                    memory = try(resources.value.requests.memory, "")
                  } : k => v if v != ""
                }
              }
            }

            dynamic "liveness_probe" {
              for_each = try(container.value.liveness_probe, null) != null ? [container.value.liveness_probe] : []
              content {
                initial_delay_seconds = try(liveness_probe.value.initial_delay_seconds, 0) > 0 ? liveness_probe.value.initial_delay_seconds : null
                period_seconds        = try(liveness_probe.value.period_seconds, 0) > 0 ? liveness_probe.value.period_seconds : null
                timeout_seconds       = try(liveness_probe.value.timeout_seconds, 0) > 0 ? liveness_probe.value.timeout_seconds : null
                success_threshold     = try(liveness_probe.value.success_threshold, 0) > 0 ? liveness_probe.value.success_threshold : null
                failure_threshold     = try(liveness_probe.value.failure_threshold, 0) > 0 ? liveness_probe.value.failure_threshold : null

                dynamic "http_get" {
                  for_each = try(liveness_probe.value.http_get, null) != null ? [liveness_probe.value.http_get] : []
                  content {
                    path   = try(http_get.value.path, "") != "" ? http_get.value.path : null
                    port   = try(http_get.value.port_number, 0) > 0 ? tostring(http_get.value.port_number) : http_get.value.port_name
                    host   = try(http_get.value.host, "") != "" ? http_get.value.host : null
                    scheme = try(http_get.value.scheme, "") != "" ? http_get.value.scheme : null
                    dynamic "http_header" {
                      for_each = try(http_get.value.http_headers, [])
                      content {
                        name  = http_header.value.name
                        value = http_header.value.value
                      }
                    }
                  }
                }
                dynamic "grpc" {
                  for_each = try(liveness_probe.value.grpc, null) != null ? [liveness_probe.value.grpc] : []
                  content {
                    port    = grpc.value.port
                    service = try(grpc.value.service, "") != "" ? grpc.value.service : null
                  }
                }
                dynamic "tcp_socket" {
                  for_each = try(liveness_probe.value.tcp_socket, null) != null ? [liveness_probe.value.tcp_socket] : []
                  content {
                    port = try(tcp_socket.value.port_number, 0) > 0 ? tostring(tcp_socket.value.port_number) : tcp_socket.value.port_name
                  }
                }
                dynamic "exec" {
                  for_each = try(liveness_probe.value.exec, null) != null ? [liveness_probe.value.exec] : []
                  content {
                    command = exec.value.command
                  }
                }
              }
            }

            dynamic "readiness_probe" {
              for_each = try(container.value.readiness_probe, null) != null ? [container.value.readiness_probe] : []
              content {
                initial_delay_seconds = try(readiness_probe.value.initial_delay_seconds, 0) > 0 ? readiness_probe.value.initial_delay_seconds : null
                period_seconds        = try(readiness_probe.value.period_seconds, 0) > 0 ? readiness_probe.value.period_seconds : null
                timeout_seconds       = try(readiness_probe.value.timeout_seconds, 0) > 0 ? readiness_probe.value.timeout_seconds : null
                success_threshold     = try(readiness_probe.value.success_threshold, 0) > 0 ? readiness_probe.value.success_threshold : null
                failure_threshold     = try(readiness_probe.value.failure_threshold, 0) > 0 ? readiness_probe.value.failure_threshold : null

                dynamic "http_get" {
                  for_each = try(readiness_probe.value.http_get, null) != null ? [readiness_probe.value.http_get] : []
                  content {
                    path   = try(http_get.value.path, "") != "" ? http_get.value.path : null
                    port   = try(http_get.value.port_number, 0) > 0 ? tostring(http_get.value.port_number) : http_get.value.port_name
                    host   = try(http_get.value.host, "") != "" ? http_get.value.host : null
                    scheme = try(http_get.value.scheme, "") != "" ? http_get.value.scheme : null
                    dynamic "http_header" {
                      for_each = try(http_get.value.http_headers, [])
                      content {
                        name  = http_header.value.name
                        value = http_header.value.value
                      }
                    }
                  }
                }
                dynamic "grpc" {
                  for_each = try(readiness_probe.value.grpc, null) != null ? [readiness_probe.value.grpc] : []
                  content {
                    port    = grpc.value.port
                    service = try(grpc.value.service, "") != "" ? grpc.value.service : null
                  }
                }
                dynamic "tcp_socket" {
                  for_each = try(readiness_probe.value.tcp_socket, null) != null ? [readiness_probe.value.tcp_socket] : []
                  content {
                    port = try(tcp_socket.value.port_number, 0) > 0 ? tostring(tcp_socket.value.port_number) : tcp_socket.value.port_name
                  }
                }
                dynamic "exec" {
                  for_each = try(readiness_probe.value.exec, null) != null ? [readiness_probe.value.exec] : []
                  content {
                    command = exec.value.command
                  }
                }
              }
            }

            dynamic "startup_probe" {
              for_each = try(container.value.startup_probe, null) != null ? [container.value.startup_probe] : []
              content {
                initial_delay_seconds = try(startup_probe.value.initial_delay_seconds, 0) > 0 ? startup_probe.value.initial_delay_seconds : null
                period_seconds        = try(startup_probe.value.period_seconds, 0) > 0 ? startup_probe.value.period_seconds : null
                timeout_seconds       = try(startup_probe.value.timeout_seconds, 0) > 0 ? startup_probe.value.timeout_seconds : null
                success_threshold     = try(startup_probe.value.success_threshold, 0) > 0 ? startup_probe.value.success_threshold : null
                failure_threshold     = try(startup_probe.value.failure_threshold, 0) > 0 ? startup_probe.value.failure_threshold : null

                dynamic "http_get" {
                  for_each = try(startup_probe.value.http_get, null) != null ? [startup_probe.value.http_get] : []
                  content {
                    path   = try(http_get.value.path, "") != "" ? http_get.value.path : null
                    port   = try(http_get.value.port_number, 0) > 0 ? tostring(http_get.value.port_number) : http_get.value.port_name
                    host   = try(http_get.value.host, "") != "" ? http_get.value.host : null
                    scheme = try(http_get.value.scheme, "") != "" ? http_get.value.scheme : null
                    dynamic "http_header" {
                      for_each = try(http_get.value.http_headers, [])
                      content {
                        name  = http_header.value.name
                        value = http_header.value.value
                      }
                    }
                  }
                }
                dynamic "grpc" {
                  for_each = try(startup_probe.value.grpc, null) != null ? [startup_probe.value.grpc] : []
                  content {
                    port    = grpc.value.port
                    service = try(grpc.value.service, "") != "" ? grpc.value.service : null
                  }
                }
                dynamic "tcp_socket" {
                  for_each = try(startup_probe.value.tcp_socket, null) != null ? [startup_probe.value.tcp_socket] : []
                  content {
                    port = try(tcp_socket.value.port_number, 0) > 0 ? tostring(tcp_socket.value.port_number) : tcp_socket.value.port_name
                  }
                }
                dynamic "exec" {
                  for_each = try(startup_probe.value.exec, null) != null ? [startup_probe.value.exec] : []
                  content {
                    command = exec.value.command
                  }
                }
              }
            }

            dynamic "volume_mount" {
              for_each = try(container.value.volume_mounts, [])
              content {
                name       = volume_mount.value.name
                mount_path = volume_mount.value.mount_path
                read_only  = try(volume_mount.value.read_only, false)
                sub_path   = try(volume_mount.value.sub_path, "") != "" ? volume_mount.value.sub_path : null
              }
            }

            # PARITY-EXCEPTION: the Terraform kubernetes provider's lifecycle
            # handlers support exec/http_get/tcp_socket but NOT the kubelet-native
            # sleep action the Pulumi module renders. A spec using the sleep hook
            # deploys identically through Pulumi; on Terraform express the same
            # drain with exec ["/bin/sleep", "N"] (requires a sleep binary in the
            # image). Stack outputs are unaffected.
            dynamic "lifecycle" {
              for_each = try(container.value.lifecycle, null) != null ? [container.value.lifecycle] : []
              content {
                dynamic "post_start" {
                  for_each = try(lifecycle.value.post_start, null) != null ? [lifecycle.value.post_start] : []
                  content {
                    dynamic "exec" {
                      for_each = try(post_start.value.exec, null) != null ? [post_start.value.exec] : []
                      content {
                        command = exec.value.command
                      }
                    }
                    dynamic "http_get" {
                      for_each = try(post_start.value.http_get, null) != null ? [post_start.value.http_get] : []
                      content {
                        path   = try(http_get.value.path, "") != "" ? http_get.value.path : null
                        port   = try(http_get.value.port_number, 0) > 0 ? tostring(http_get.value.port_number) : http_get.value.port_name
                        scheme = try(http_get.value.scheme, "") != "" ? http_get.value.scheme : null
                      }
                    }
                    dynamic "tcp_socket" {
                      for_each = try(post_start.value.tcp_socket, null) != null ? [post_start.value.tcp_socket] : []
                      content {
                        port = try(tcp_socket.value.port_number, 0) > 0 ? tostring(tcp_socket.value.port_number) : tcp_socket.value.port_name
                      }
                    }
                  }
                }
                dynamic "pre_stop" {
                  for_each = try(lifecycle.value.pre_stop, null) != null ? [lifecycle.value.pre_stop] : []
                  content {
                    dynamic "exec" {
                      for_each = try(pre_stop.value.exec, null) != null ? [pre_stop.value.exec] : []
                      content {
                        command = exec.value.command
                      }
                    }
                    dynamic "http_get" {
                      for_each = try(pre_stop.value.http_get, null) != null ? [pre_stop.value.http_get] : []
                      content {
                        path   = try(http_get.value.path, "") != "" ? http_get.value.path : null
                        port   = try(http_get.value.port_number, 0) > 0 ? tostring(http_get.value.port_number) : http_get.value.port_name
                        scheme = try(http_get.value.scheme, "") != "" ? http_get.value.scheme : null
                      }
                    }
                    dynamic "tcp_socket" {
                      for_each = try(pre_stop.value.tcp_socket, null) != null ? [pre_stop.value.tcp_socket] : []
                      content {
                        port = try(tcp_socket.value.port_number, 0) > 0 ? tostring(tcp_socket.value.port_number) : tcp_socket.value.port_name
                      }
                    }
                  }
                }
              }
            }

            dynamic "security_context" {
              for_each = try(container.value.security_context, null) != null ? [container.value.security_context] : []
              content {
                privileged                 = try(security_context.value.privileged, false)
                run_as_user                = try(security_context.value.run_as_user, null)
                run_as_group               = try(security_context.value.run_as_group, null)
                run_as_non_root            = try(security_context.value.run_as_non_root, null)
                read_only_root_filesystem  = try(security_context.value.read_only_root_filesystem, null)
                allow_privilege_escalation = try(security_context.value.allow_privilege_escalation, null)
                dynamic "capabilities" {
                  for_each = try(security_context.value.capabilities, null) != null ? [security_context.value.capabilities] : []
                  content {
                    add  = length(try(capabilities.value.add, [])) > 0 ? capabilities.value.add : null
                    drop = length(try(capabilities.value.drop, [])) > 0 ? capabilities.value.drop : null
                  }
                }
                dynamic "seccomp_profile" {
                  for_each = try(security_context.value.seccomp_profile, null) != null ? [security_context.value.seccomp_profile] : []
                  content {
                    type              = seccomp_profile.value.type
                    localhost_profile = try(seccomp_profile.value.localhost_profile, "") != "" ? seccomp_profile.value.localhost_profile : null
                  }
                }
              }
            }
          }
        }

        # Pod volumes: union of every container's mounts, de-duplicated by name
        # (first declaration wins — see locals.pod_volumes).
        dynamic "volume" {
          for_each = local.pod_volumes
          content {
            name = volume.key

            dynamic "config_map" {
              for_each = try(volume.value[0].config_map, null) != null ? [volume.value[0].config_map] : []
              content {
                name         = config_map.value.name
                default_mode = try(config_map.value.default_mode, 0) > 0 ? format("%04o", config_map.value.default_mode) : null
                dynamic "items" {
                  for_each = try(config_map.value.key, "") != "" ? [config_map.value] : []
                  content {
                    key  = items.value.key
                    path = try(items.value.path, "") != "" ? items.value.path : items.value.key
                  }
                }
              }
            }

            dynamic "secret" {
              for_each = try(volume.value[0].secret, null) != null ? [volume.value[0].secret] : []
              content {
                secret_name  = secret.value.name
                default_mode = try(secret.value.default_mode, 0) > 0 ? format("%04o", secret.value.default_mode) : null
                dynamic "items" {
                  for_each = try(secret.value.key, "") != "" ? [secret.value] : []
                  content {
                    key  = items.value.key
                    path = try(items.value.path, "") != "" ? items.value.path : items.value.key
                  }
                }
              }
            }

            dynamic "host_path" {
              for_each = try(volume.value[0].host_path, null) != null ? [volume.value[0].host_path] : []
              content {
                path = host_path.value.path
                type = try(host_path.value.type, "") != "" ? host_path.value.type : null
              }
            }

            dynamic "empty_dir" {
              for_each = try(volume.value[0].empty_dir, null) != null ? [volume.value[0].empty_dir] : []
              content {
                medium     = try(empty_dir.value.medium, "") != "" ? empty_dir.value.medium : null
                size_limit = try(empty_dir.value.size_limit, "") != "" ? empty_dir.value.size_limit : null
              }
            }

            dynamic "persistent_volume_claim" {
              for_each = try(volume.value[0].pvc, null) != null ? [volume.value[0].pvc] : []
              content {
                claim_name = persistent_volume_claim.value.claim_name
                read_only  = try(persistent_volume_claim.value.read_only, false)
              }
            }

            # Kubelet-rotated, audience-bound ServiceAccount tokens ride the
            # "projected" volume type in the Kubernetes API.
            dynamic "projected" {
              for_each = try(volume.value[0].service_account_token, null) != null ? [volume.value[0].service_account_token] : []
              content {
                sources {
                  service_account_token {
                    audience           = projected.value.audience
                    expiration_seconds = try(projected.value.expiration_seconds, 0) > 0 ? projected.value.expiration_seconds : null
                    path               = try(projected.value.path, "") != "" ? projected.value.path : "token"
                  }
                }
              }
            }
          }
        }
      }
    }
  }

  depends_on = [kubernetes_namespace.this]
}
