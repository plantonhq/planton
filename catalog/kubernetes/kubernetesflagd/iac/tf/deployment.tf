# The flagd Deployment (Pulumi twin: deployment.go). ConfigMap sources are
# mounted as DIRECTORIES (never subPath, which would never see an edit);
# flagd reads its sources from FLAGD_SOURCES and every other setting from its
# arguments. Liveness is /healthz and readiness /readyz on the management
# port - flagd reports ready only once every source has synced. The sources
# checksum on the pod template rolls the pods when a source changes.
resource "kubernetes_deployment_v1" "flagd" {
  metadata {
    name      = local.name
    namespace = local.namespace
    labels    = local.labels
  }

  spec {
    replicas = local.replicas

    selector {
      match_labels = local.selector_labels
    }

    template {
      metadata {
        labels      = local.pod_labels
        annotations = merge({ "checksum/sources" = local.sources_checksum }, try(var.spec.pod_annotations, {}))
      }

      spec {
        service_account_name = local.service_account_name
        node_selector        = length(try(var.spec.scheduling.node_selector, {})) > 0 ? var.spec.scheduling.node_selector : null
        scheduler_name       = try(var.spec.scheduling.scheduler_name, "") != "" ? var.spec.scheduling.scheduler_name : null

        dynamic "image_pull_secrets" {
          for_each = try(var.spec.image.pull_secret_names, [])
          content {
            name = image_pull_secrets.value
          }
        }

        dynamic "toleration" {
          for_each = try(var.spec.scheduling.tolerations, [])
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
            try(var.spec.scheduling.node_affinity, null) != null ||
            try(var.spec.scheduling.pod_affinity, null) != null ||
            try(var.spec.scheduling.pod_anti_affinity, null) != null
          ) ? [var.spec.scheduling] : []
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
          for_each = local.topology_spread_constraints
          content {
            max_skew           = topology_spread_constraint.value.max_skew
            topology_key       = topology_spread_constraint.value.topology_key
            when_unsatisfiable = topology_spread_constraint.value.when_unsatisfiable
            label_selector {
              match_labels = topology_spread_constraint.value.match_labels
            }
          }
        }

        dynamic "security_context" {
          for_each = try(var.spec.pod_security_context, null) != null ? [var.spec.pod_security_context] : []
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

        container {
          name              = "flagd"
          image             = local.image
          image_pull_policy = local.pull_policy
          args              = local.args

          env {
            name = "FLAGD_SOURCES"
            value_from {
              secret_key_ref {
                name = kubernetes_secret_v1.sources.metadata[0].name
                key  = "sources"
              }
            }
          }

          dynamic "env" {
            for_each = { for k in sort(keys(try(var.spec.extra_env, {}))) : k => var.spec.extra_env[k] }
            content {
              name  = env.key
              value = env.value
            }
          }

          dynamic "env" {
            for_each = { for k in sort(keys(try(var.spec.extra_env_from_secret, {}))) : k => var.spec.extra_env_from_secret[k] }
            content {
              name = env.key
              value_from {
                secret_key_ref {
                  name = env.value.name
                  key  = env.value.key
                }
              }
            }
          }

          port {
            name           = "evaluation"
            container_port = local.port
          }
          port {
            name           = "management"
            container_port = local.management_port
          }
          port {
            name           = "sync"
            container_port = local.sync_port
          }
          port {
            name           = "ofrep"
            container_port = local.ofrep_port
          }

          liveness_probe {
            http_get {
              path = "/healthz"
              port = "management"
            }
            initial_delay_seconds = 5
            period_seconds        = 10
          }

          readiness_probe {
            http_get {
              path = "/readyz"
              port = "management"
            }
            initial_delay_seconds = 5
            period_seconds        = 5
          }

          dynamic "resources" {
            for_each = try(var.spec.resources, null) != null ? [var.spec.resources] : []
            content {
              requests = { for k, v in { cpu = try(resources.value.requests.cpu, ""), memory = try(resources.value.requests.memory, "") } : k => v if v != "" && v != null }
              limits   = { for k, v in { cpu = try(resources.value.limits.cpu, ""), memory = try(resources.value.limits.memory, "") } : k => v if v != "" && v != null }
            }
          }

          dynamic "security_context" {
            for_each = try(var.spec.container_security_context, null) != null ? [var.spec.container_security_context] : []
            content {
              privileged                 = try(security_context.value.privileged, false) == true ? true : null
              run_as_user                = try(security_context.value.run_as_user, null)
              run_as_group               = try(security_context.value.run_as_group, null)
              run_as_non_root            = try(security_context.value.run_as_non_root, null)
              read_only_root_filesystem  = try(security_context.value.read_only_root_filesystem, null)
              allow_privilege_escalation = try(security_context.value.allow_privilege_escalation, null)
              dynamic "capabilities" {
                for_each = (length(try(security_context.value.capabilities.add, [])) + length(try(security_context.value.capabilities.drop, []))) > 0 ? [security_context.value.capabilities] : []
                content {
                  add  = try(capabilities.value.add, [])
                  drop = try(capabilities.value.drop, [])
                }
              }
              dynamic "seccomp_profile" {
                for_each = try(security_context.value.seccomp_profile.type, "") != "" ? [security_context.value.seccomp_profile] : []
                content {
                  type              = seccomp_profile.value.type
                  localhost_profile = try(seccomp_profile.value.localhost_profile, "") != "" ? seccomp_profile.value.localhost_profile : null
                }
              }
            }
          }

          dynamic "volume_mount" {
            for_each = local.config_map_mounts
            content {
              name       = volume_mount.value.volume
              mount_path = volume_mount.value.mount_path
              read_only  = true
            }
          }
          dynamic "volume_mount" {
            for_each = local.grpc_ca_mounts
            content {
              name       = volume_mount.value.volume
              mount_path = volume_mount.value.mount_path
              read_only  = true
            }
          }
          dynamic "volume_mount" {
            for_each = local.tls_secret_name != "" ? [1] : []
            content {
              name       = "server-tls"
              mount_path = "/etc/flagd/tls"
              read_only  = true
            }
          }
          dynamic "volume_mount" {
            for_each = local.otel_ca != null ? [1] : []
            content {
              name       = "otel-ca"
              mount_path = "/etc/flagd/otel-ca"
              read_only  = true
            }
          }
          dynamic "volume_mount" {
            for_each = local.otel_client_tls_secret != "" ? [1] : []
            content {
              name       = "otel-tls"
              mount_path = "/etc/flagd/otel-tls"
              read_only  = true
            }
          }
        }

        dynamic "volume" {
          for_each = local.config_map_mounts
          content {
            name = volume.value.volume
            config_map {
              name = volume.value.config_map
            }
          }
        }
        dynamic "volume" {
          for_each = local.grpc_ca_mounts
          content {
            name = volume.value.volume
            secret {
              secret_name = volume.value.secret
              items {
                key  = volume.value.key
                path = volume.value.key
              }
            }
          }
        }
        dynamic "volume" {
          for_each = local.tls_secret_name != "" ? [1] : []
          content {
            name = "server-tls"
            secret {
              secret_name = local.tls_secret_name
            }
          }
        }
        dynamic "volume" {
          for_each = local.otel_ca != null ? [local.otel_ca] : []
          content {
            name = "otel-ca"
            secret {
              secret_name = volume.value.name
              items {
                key  = volume.value.key
                path = volume.value.key
              }
            }
          }
        }
        dynamic "volume" {
          for_each = local.otel_client_tls_secret != "" ? [1] : []
          content {
            name = "otel-tls"
            secret {
              secret_name = local.otel_client_tls_secret
            }
          }
        }
      }
    }
  }

  depends_on = [
    kubernetes_namespace_v1.flagd,
    kubernetes_secret_v1.sources,
    kubernetes_service_account_v1.flagd,
    kubernetes_role_binding_v1.flag_reader,
  ]

  lifecycle {
    precondition {
      condition     = length(var.metadata.name) <= 63
      error_message = "metadata.name exceeds flagd's 63-character name budget (the Service is named after the resource, and a Service name is a 63-character DNS label)."
    }
    precondition {
      condition     = length(local.header_conflicts) == 0
      error_message = "A header is declared both plain and sensitive; declare it once."
    }
  }
}
