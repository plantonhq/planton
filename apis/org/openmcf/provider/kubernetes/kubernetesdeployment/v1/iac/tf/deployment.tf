# 1) Create a ServiceAccount
resource "kubernetes_service_account" "this" {
  metadata {
    name      = local.resource_id
    namespace = local.namespace
  }
}

# 3) Create the Deployment
resource "kubernetes_deployment" "this" {
  metadata {
    name      = var.metadata.name
    namespace = local.namespace
    labels    = local.final_labels
    annotations = {
      # Example annotation (remove or modify as needed)
      "example.annotation" = "true"
    }
  }

  spec {
    # If not provided, default to 1 replica
    replicas = try(var.spec.availability.min_replicas, 1)

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
          name  = "microservice"
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

          # Add env variables from var.spec.container.app.env.variables
          # After generator flattening, each value is a plain string.
          dynamic "env" {
            for_each = {
              for k, v in (try(var.spec.container.app.env.variables, null) != null ? var.spec.container.app.env.variables : {}) :
              k => v
              if v != null && v != ""
            }
            content {
              name  = env.key
              value = env.value
            }
          }

          # Add env variables from secrets with direct string values (using computed secret name)
          dynamic "env" {
            for_each = {
              for k, v in (try(var.spec.container.app.env.secrets, null) != null ? var.spec.container.app.env.secrets : {}) :
              k => v
              if try(v.value, null) != null && v.value != ""
            }
            content {
              name = env.key
              value_from {
                secret_key_ref {
                  name = local.env_secret_name
                  key  = env.key
                }
              }
            }
          }

          # Add env variables from external Kubernetes Secret references
          dynamic "env" {
            for_each = {
              for k, v in (try(var.spec.container.app.env.secrets, null) != null ? var.spec.container.app.env.secrets : {}) :
              k => v
              if try(v.secret_ref, null) != null
            }
            content {
              name = env.key
              value_from {
                secret_key_ref {
                  name = env.value.secret_ref.name
                  key  = env.value.secret_ref.key
                }
              }
            }
          }

          # Resource requests/limits
          resources {
            limits = {
              cpu = try(var.spec.container.app.resources.limits.cpu, null)
              memory = try(var.spec.container.app.resources.limits.memory, null)
            }
            requests = {
              cpu = try(var.spec.container.app.resources.requests.cpu, null)
              memory = try(var.spec.container.app.resources.requests.memory, null)
            }
          }

          # Volume mounts for the container
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
          command = length(try(var.spec.container.app.command, [])) > 0 ? var.spec.container.app.command : null

          # Args override (if specified)
          args = length(try(var.spec.container.app.args, [])) > 0 ? var.spec.container.app.args : null

          # Lifecycle pre-stop hook (sleep 60 seconds)
          lifecycle {
            pre_stop {
              exec {
                command = ["/bin/sleep", "60"]
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
    kubernetes_config_map.this
  ]
}
