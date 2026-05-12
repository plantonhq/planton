##############################################
# daemonset.tf
#
# Creates the Kubernetes DaemonSet with:
#  - Main application container
#  - Optional sidecar containers
#  - Security context for privileged operations
#  - Environment variables from direct values and secrets
#  - Environment secrets from both string values and external secret refs
#  - Volume mounts (HostPath, ConfigMap, Secret, EmptyDir, PVC)
#  - Tolerations for scheduling on tainted nodes
#  - Node selector for targeting specific nodes
#  - Update strategy (RollingUpdate or OnDelete)
##############################################

resource "kubernetes_daemon_set_v1" "this" {
  metadata {
    name      = var.metadata.name
    namespace = local.namespace
    labels    = local.final_labels
  }

  spec {
    selector {
      match_labels = local.selector_labels
    }

    # Min ready seconds
    min_ready_seconds = var.spec.min_ready_seconds

    # Update strategy
    dynamic "strategy" {
      for_each = var.spec.update_strategy != null ? [var.spec.update_strategy] : []
      content {
        type = strategy.value.type
        dynamic "rolling_update" {
          for_each = strategy.value.type == "RollingUpdate" && strategy.value.rolling_update != null ? [strategy.value.rolling_update] : []
          content {
            max_unavailable = try(rolling_update.value.max_unavailable, null)
            max_surge       = try(rolling_update.value.max_surge, null)
          }
        }
      }
    }

    template {
      metadata {
        labels = local.final_labels
      }

      spec {
        # ServiceAccount
        service_account_name = var.spec.create_service_account ? kubernetes_service_account.this[0].metadata[0].name : try(var.spec.service_account_name, null)

        # Node selector
        node_selector = var.spec.node_selector

        # Tolerations
        dynamic "toleration" {
          for_each = var.spec.tolerations
          content {
            key                = try(toleration.value.key, null)
            operator           = try(toleration.value.operator, null)
            value              = try(toleration.value.value, null)
            effect             = try(toleration.value.effect, null)
            toleration_seconds = try(toleration.value.toleration_seconds, null)
          }
        }

        # Image pull secrets
        dynamic "image_pull_secrets" {
          for_each = var.spec.container.app.image.pull_secret_name != null ? [1] : []
          content {
            name = var.spec.container.app.image.pull_secret_name
          }
        }

        # Main container
        container {
          name  = "daemonset-container"
          image = "${var.spec.container.app.image.repo}:${var.spec.container.app.image.tag}"

          # Container ports
          dynamic "port" {
            for_each = try(var.spec.container.app.ports, [])
            content {
              name           = port.value.name
              container_port = port.value.container_port
              protocol       = port.value.network_protocol
              host_port      = try(port.value.host_port, null)
            }
          }

          # Built-in environment variables
          env {
            name = "HOSTNAME"
            value_from {
              field_ref {
                field_path = "status.podIP"
              }
            }
          }

          env {
            name = "K8S_POD_ID"
            value_from {
              field_ref {
                field_path  = "metadata.name"
                api_version = "v1"
              }
            }
          }

          env {
            name = "K8S_NODE_NAME"
            value_from {
              field_ref {
                field_path = "spec.nodeName"
              }
            }
          }

          # Environment variables with direct values
          dynamic "env" {
            for_each = [for v in try(var.spec.container.app.env.variables, []) : v if v.value != null]
            content {
              name  = env.value.name
              value = env.value.value
            }
          }

          # Environment variables from ConfigMap key references
          dynamic "env" {
            for_each = [for v in try(var.spec.container.app.env.variables, []) : v if v.config_map_key_ref != null]
            content {
              name = env.value.name
              value_from {
                config_map_key_ref {
                  name     = env.value.config_map_key_ref.name
                  key      = env.value.config_map_key_ref.key
                  optional = try(env.value.config_map_key_ref.optional, false)
                }
              }
            }
          }

          # Environment variables from pod field references
          dynamic "env" {
            for_each = [for v in try(var.spec.container.app.env.variables, []) : v if v.field_ref != null]
            content {
              name = env.value.name
              value_from {
                field_ref {
                  field_path  = env.value.field_ref.field_path
                  api_version = try(env.value.field_ref.api_version, null)
                }
              }
            }
          }

          # Environment variables from container resource field references
          dynamic "env" {
            for_each = [for v in try(var.spec.container.app.env.variables, []) : v if v.resource_field_ref != null]
            content {
              name = env.value.name
              value_from {
                resource_field_ref {
                  resource       = env.value.resource_field_ref.resource
                  container_name = try(env.value.resource_field_ref.container_name, null)
                  divisor        = try(env.value.resource_field_ref.divisor, null)
                }
              }
            }
          }

          # Secret environment variables with direct string values (auto-created K8s Secret)
          dynamic "env" {
            for_each = [for s in try(var.spec.container.app.env.secrets, []) : s if try(s.value, null) != null && s.value != ""]
            content {
              name = env.value.name
              value_from {
                secret_key_ref {
                  name = local.env_secret_name
                  key  = env.value.name
                }
              }
            }
          }

          # Secret environment variables from external Kubernetes Secret references
          dynamic "env" {
            for_each = [for s in try(var.spec.container.app.env.secrets, []) : s if s.secret_ref != null]
            content {
              name = env.value.name
              value_from {
                secret_key_ref {
                  name     = env.value.secret_ref.name
                  key      = env.value.secret_ref.key
                  optional = try(env.value.secret_ref.optional, false)
                }
              }
            }
          }

          # Bulk envFrom: import all keys from ConfigMaps or Secrets
          dynamic "env_from" {
            for_each = try(var.spec.container.app.env.env_from, [])
            content {
              prefix = try(env_from.value.prefix, null)
              dynamic "config_map_ref" {
                for_each = env_from.value.config_map_ref != null ? [env_from.value.config_map_ref] : []
                content {
                  name     = config_map_ref.value.name
                  optional = try(config_map_ref.value.optional, false)
                }
              }
              dynamic "secret_ref" {
                for_each = env_from.value.secret_ref != null ? [env_from.value.secret_ref] : []
                content {
                  name     = secret_ref.value.name
                  optional = try(secret_ref.value.optional, false)
                }
              }
            }
          }

          # Resource requests/limits
          resources {
            limits = {
              cpu    = try(var.spec.container.app.resources.limits.cpu, null)
              memory = try(var.spec.container.app.resources.limits.memory, null)
            }
            requests = {
              cpu    = try(var.spec.container.app.resources.requests.cpu, null)
              memory = try(var.spec.container.app.resources.requests.memory, null)
            }
          }

          # Volume mounts
          dynamic "volume_mount" {
            for_each = try(var.spec.container.app.volume_mounts, [])
            content {
              name       = volume_mount.value.name
              mount_path = volume_mount.value.mount_path
              read_only  = try(volume_mount.value.read_only, false)
              sub_path   = try(volume_mount.value.sub_path, null)
            }
          }

          # Command override
          command = length(try(var.spec.container.app.command, [])) > 0 ? var.spec.container.app.command : null

          # Args override
          args = length(try(var.spec.container.app.args, [])) > 0 ? var.spec.container.app.args : null

          # Security context
          dynamic "security_context" {
            for_each = var.spec.container.app.security_context != null ? [var.spec.container.app.security_context] : []
            content {
              privileged                = try(security_context.value.privileged, false)
              run_as_user               = try(security_context.value.run_as_user, null)
              run_as_group              = try(security_context.value.run_as_group, null)
              run_as_non_root           = try(security_context.value.run_as_non_root, null)
              read_only_root_filesystem = try(security_context.value.read_only_root_filesystem, false)

              dynamic "capabilities" {
                for_each = security_context.value.capabilities != null ? [security_context.value.capabilities] : []
                content {
                  add  = try(capabilities.value.add, [])
                  drop = try(capabilities.value.drop, [])
                }
              }
            }
          }
        }

        # Sidecar containers
        dynamic "container" {
          for_each = try(var.spec.container.sidecars, [])
          content {
            name  = container.value.name
            image = container.value.image

            dynamic "port" {
              for_each = try(container.value.ports, [])
              content {
                name           = port.value.name
                container_port = port.value.container_port
                protocol       = port.value.protocol
              }
            }

            resources {
              limits = {
                cpu    = try(container.value.resources.limits.cpu, null)
                memory = try(container.value.resources.limits.memory, null)
              }
              requests = {
                cpu    = try(container.value.resources.requests.cpu, null)
                memory = try(container.value.resources.requests.memory, null)
              }
            }

            dynamic "env" {
              for_each = try(container.value.env, [])
              content {
                name  = env.value.name
                value = env.value.value
              }
            }
          }
        }

        # ConfigMap volumes
        dynamic "volume" {
          for_each = [for vm in try(var.spec.container.app.volume_mounts, []) : vm if vm.config_map != null]
          content {
            name = volume.value.name
            config_map {
              name         = volume.value.config_map.name
              default_mode = try(volume.value.config_map.default_mode, null)
              dynamic "items" {
                for_each = volume.value.config_map.key != null ? [1] : []
                content {
                  key  = volume.value.config_map.key
                  path = try(volume.value.config_map.path, volume.value.config_map.key)
                }
              }
            }
          }
        }

        # Secret volumes
        dynamic "volume" {
          for_each = [for vm in try(var.spec.container.app.volume_mounts, []) : vm if vm.secret != null]
          content {
            name = volume.value.name
            secret {
              secret_name  = volume.value.secret.name
              default_mode = try(volume.value.secret.default_mode, null)
              dynamic "items" {
                for_each = volume.value.secret.key != null ? [1] : []
                content {
                  key  = volume.value.secret.key
                  path = try(volume.value.secret.path, volume.value.secret.key)
                }
              }
            }
          }
        }

        # HostPath volumes (common for DaemonSets)
        dynamic "volume" {
          for_each = [for vm in try(var.spec.container.app.volume_mounts, []) : vm if vm.host_path != null]
          content {
            name = volume.value.name
            host_path {
              path = volume.value.host_path.path
              type = try(volume.value.host_path.type, null)
            }
          }
        }

        # EmptyDir volumes
        dynamic "volume" {
          for_each = [for vm in try(var.spec.container.app.volume_mounts, []) : vm if vm.empty_dir != null]
          content {
            name = volume.value.name
            empty_dir {
              medium     = try(volume.value.empty_dir.medium, null)
              size_limit = try(volume.value.empty_dir.size_limit, null)
            }
          }
        }

        # PVC volumes
        dynamic "volume" {
          for_each = [for vm in try(var.spec.container.app.volume_mounts, []) : vm if vm.pvc != null]
          content {
            name = volume.value.name
            persistent_volume_claim {
              claim_name = volume.value.pvc.claim_name
              read_only  = try(volume.value.pvc.read_only, false)
            }
          }
        }
      }
    }
  }

  depends_on = [
    kubernetes_namespace.this,
    kubernetes_config_map.this,
    kubernetes_secret.this
  ]
}

