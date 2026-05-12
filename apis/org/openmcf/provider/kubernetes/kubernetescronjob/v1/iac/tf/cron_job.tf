##############################################
# main.tf
##############################################

# 1) Create a ServiceAccount
resource "kubernetes_service_account" "this" {
  metadata {
    name      = local.resource_id
    namespace = local.namespace_name
  }

  depends_on = [
    kubernetes_namespace.this
  ]
}

# 2) Create an optional image pull secret if Docker credentials are provided
resource "kubernetes_secret" "image_pull_secret" {
  count = var.docker_config_json != "" ? 1 : 0

  metadata {
    name      = local.image_pull_secret_name
    namespace = local.namespace_name
    labels    = local.final_labels
  }
  type = "kubernetes.io/dockerconfigjson"

  data = {
    ".dockerconfigjson" = var.docker_config_json
  }

  depends_on = [
    kubernetes_namespace.this
  ]
}

# 3) Create the CronJob
resource "kubernetes_cron_job_v1" "this" {
  metadata {
    name      = var.metadata.name
    namespace = local.namespace_name
    labels    = local.final_labels
  }

  spec {
    schedule = var.spec.schedule

    # If not set, concurrencyPolicy defaults to "Forbid"
    concurrency_policy = try(var.spec.concurrency_policy, "Forbid")

    # If not set, default to false
    suspend = try(var.spec.suspend, false)

    # Defaults if not set
    successful_jobs_history_limit = try(var.spec.successful_jobs_history_limit, 3)
    failed_jobs_history_limit     = try(var.spec.failed_jobs_history_limit, 1)

    # For starting_deadline_seconds, if not set or zero, we use null so that it doesn't appear
    starting_deadline_seconds = (try(var.spec.starting_deadline_seconds, 0) != 0 ? var.spec.starting_deadline_seconds :
    null)

    job_template {
      metadata {
        labels = local.final_labels
      }

      spec {
        backoff_limit = try(var.spec.backoff_limit, 6)

        template {
          metadata {
            labels = local.final_labels
          }

          spec {
            # Typically "Never" is the recommended default for CronJobs
            restart_policy       = try(var.spec.restart_policy, "Never")
            service_account_name = kubernetes_service_account.this.metadata[0].name

            container {
              name  = "cronjob-container"
              image = "${var.spec.image.repo}:${var.spec.image.tag}"

              # Use the custom command and args if provided, otherwise default to empty lists
              command = try(var.spec.command, [])
              args    = try(var.spec.args, [])

              # Env variables
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

              # Environment variables with direct values
              dynamic "env" {
                for_each = [for v in try(var.spec.env.variables, []) : v if v.value != null]
                content {
                  name  = env.value.name
                  value = env.value.value
                }
              }

              # Environment variables from ConfigMap key references
              dynamic "env" {
                for_each = [for v in try(var.spec.env.variables, []) : v if v.config_map_key_ref != null]
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
                for_each = [for v in try(var.spec.env.variables, []) : v if v.field_ref != null]
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
                for_each = [for v in try(var.spec.env.variables, []) : v if v.resource_field_ref != null]
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
                for_each = [for s in try(var.spec.env.secrets, []) : s if try(s.value, null) != null && s.value != ""]
                content {
                  name = env.value.name
                  value_from {
                    secret_key_ref {
                      name = local.env_secrets_secret_name
                      key  = env.value.name
                    }
                  }
                }
              }

              # Secret environment variables from external Kubernetes Secret references
              dynamic "env" {
                for_each = [for s in try(var.spec.env.secrets, []) : s if s.secret_ref != null]
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
                for_each = try(var.spec.env.env_from, [])
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
                  cpu    = try(var.spec.resources.limits.cpu, null)
                  memory = try(var.spec.resources.limits.memory, null)
                }
                requests = {
                  cpu    = try(var.spec.resources.requests.cpu, null)
                  memory = try(var.spec.resources.requests.memory, null)
                }
              }

              # Volume mounts for the container
              dynamic "volume_mount" {
                for_each = try(var.spec.volume_mounts, [])
                content {
                  name       = volume_mount.value.name
                  mount_path = volume_mount.value.mount_path
                  read_only  = try(volume_mount.value.read_only, false)
                  sub_path   = try(volume_mount.value.sub_path, null)
                }
              }
            }

            # ConfigMap volumes
            dynamic "volume" {
              for_each = [for vm in try(var.spec.volume_mounts, []) : vm if vm.config_map != null]
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
              for_each = [for vm in try(var.spec.volume_mounts, []) : vm if vm.secret != null]
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
              for_each = [for vm in try(var.spec.volume_mounts, []) : vm if vm.host_path != null]
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
              for_each = [for vm in try(var.spec.volume_mounts, []) : vm if vm.empty_dir != null]
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
              for_each = [for vm in try(var.spec.volume_mounts, []) : vm if vm.pvc != null]
              content {
                name = volume.value.name
                persistent_volume_claim {
                  claim_name = volume.value.pvc.claim_name
                  read_only  = try(volume.value.pvc.read_only, false)
                }
              }
            }

            # Attach image pull secret if Docker credentials were provided
            dynamic "image_pull_secrets" {
              for_each = kubernetes_secret.image_pull_secret
              content {
                name = image_pull_secrets.value.metadata[0].name
              }
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
