# 1) Create a ServiceAccount
resource "kubernetes_service_account" "this" {
  metadata {
    name      = local.resource_id
    namespace = local.namespace
  }
}

# 2) Create the StatefulSet
resource "kubernetes_stateful_set" "this" {
  metadata {
    name      = var.metadata.name
    namespace = local.namespace
    labels    = local.final_labels
  }

  spec {
    service_name          = local.headless_service_name
    replicas              = local.replicas
    pod_management_policy = local.pod_management_policy

    selector {
      match_labels = local.selector_labels
    }

    template {
      metadata {
        labels = local.final_labels
      }

      spec {
        service_account_name             = kubernetes_service_account.this.metadata[0].name
        termination_grace_period_seconds = 60

        container {
          name  = "app"
          image = "${var.spec.container.app.image.repo}:${var.spec.container.app.image.tag}"

          # Container ports
          dynamic "port" {
            for_each = try(var.spec.container.app.ports, [])
            content {
              name           = port.value.name
              container_port = port.value.container_port
            }
          }

          # Add built-in environment variables
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
            name = "K8S_POD_NAMESPACE"
            value_from {
              field_ref {
                field_path  = "metadata.namespace"
                api_version = "v1"
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

          # Command override (if specified)
          command = try(length(var.spec.container.app.command), 0) > 0 ? var.spec.container.app.command : null

          # Args (if specified)
          args = try(length(var.spec.container.app.args), 0) > 0 ? var.spec.container.app.args : null
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

        # HostPath volumes
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

        # PVC volumes (only for external PVCs, not volumeClaimTemplates)
        # For StatefulSets, volumeClaimTemplate references are handled automatically
        dynamic "volume" {
          for_each = [
            for vm in try(var.spec.container.app.volume_mounts, []) : vm
            if vm.pvc != null && !contains(
              [for vct in(var.spec.volume_claim_templates != null ? var.spec.volume_claim_templates : []) : vct.name],
              vm.pvc.claim_name
            )
          ]
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

    # Volume claim templates for persistent storage
    dynamic "volume_claim_template" {
      for_each = var.spec.volume_claim_templates != null ? var.spec.volume_claim_templates : []
      content {
        metadata {
          name = volume_claim_template.value.name
        }
        spec {
          access_modes       = try(volume_claim_template.value.access_modes, ["ReadWriteOnce"])
          storage_class_name = try(volume_claim_template.value.storage_class, null)
          resources {
            requests = {
              storage = volume_claim_template.value.size
            }
          }
        }
      }
    }
  }

  depends_on = [
    kubernetes_namespace.this,
    kubernetes_service.headless,
    kubernetes_config_map.this
  ]
}
