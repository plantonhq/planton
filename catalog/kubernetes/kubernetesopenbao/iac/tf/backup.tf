# Backup and restore — module-owned, beside the release (never inside
# it): the chart's own snapshot agent is S3-only, static-keys-only, and
# cannot restore, so the module renders its own CronJob and Job on the
# official openbao and rclone images. Exact twin of the Pulumi module's
# backup.go / restore.go. The release does not wait on these and they do
# not wait on the release: a CronJob whose first run fails its login until
# the operator runs the recipe is the designed day-1 shape, taught by the
# run's own log (scripts.tf).

# The job's ServiceAccount. Its annotation is the cloud-side half of the
# keyless posture; the other half is the binding the identity kind owns
# (a GcpGkeWorkloadIdentityBinding naming this ServiceAccount).
resource "kubernetes_service_account_v1" "backup" {
  count = local.backup_enabled ? 1 : 0

  metadata {
    name        = local.backup_name
    namespace   = local.namespace
    labels      = local.labels
    annotations = length(local.backup_sa_annotations) > 0 ? local.backup_sa_annotations : null
  }

  depends_on = [kubernetes_namespace_v1.openbao]
}

# The four scripts, byte-identical to the Pulumi module's constants.
resource "kubernetes_config_map_v1" "backup_scripts" {
  count = local.backup_enabled ? 1 : 0

  metadata {
    name      = local.backup_scripts_name
    namespace = local.namespace
    labels    = local.labels
  }

  data = local.backup_scripts

  depends_on = [kubernetes_namespace_v1.openbao]
}

# Declared store credentials, keyed by the rclone environment variable
# each reaches the job as (plus the ca.pem file for a self-signed S3
# endpoint). Absent for the keyless postures. Credential material never
# rides a ConfigMap or a rendered pod spec.
resource "kubernetes_secret_v1" "backup_credentials" {
  count = local.backup_enabled && local.backup_has_credentials ? 1 : 0

  metadata {
    name      = local.backup_credentials_name
    namespace = local.namespace
    labels    = local.labels
  }

  data = local.backup_credentials_data

  depends_on = [kubernetes_namespace_v1.openbao]
}

# The scheduled snapshot: an init container takes the snapshot with the
# bao CLI, the main container ships and prunes with rclone. One run at a
# time; a run that outlives the deadline is a failure, not a wait.
# SUSPENDED while `restore` is declared: a fresh target must not snapshot
# an empty vault into the shared prefix (and prune the source's
# snapshots) while its restore Job waits for the operator's token.
resource "kubernetes_cron_job_v1" "backup" {
  count = local.backup_enabled ? 1 : 0

  metadata {
    name      = local.backup_name
    namespace = local.namespace
    labels    = local.labels
  }

  spec {
    schedule                      = local.backup_schedule
    concurrency_policy            = "Forbid"
    suspend                       = local.restore_declared
    successful_jobs_history_limit = local.backup_jobs_history
    failed_jobs_history_limit     = local.backup_jobs_history

    job_template {
      metadata {}
      spec {
        backoff_limit           = local.backup_job_backoff_limit
        active_deadline_seconds = local.backup_job_deadline_sec

        template {
          metadata {
            labels = local.backup_pod_labels
          }
          spec {
            service_account_name = local.backup_name
            restart_policy       = "OnFailure"
            node_selector        = length(try(var.spec.server.scheduling.node_selector, {})) > 0 ? var.spec.server.scheduling.node_selector : null

            # OpenBao's image runs as uid 100; the rclone image declares
            # no user. One identity for both containers so the 0600
            # snapshot file the bao CLI writes is readable by rclone.
            security_context {
              run_as_user     = local.backup_job_run_as_user
              run_as_group    = local.backup_job_run_as_group
              fs_group        = local.backup_job_run_as_group
              run_as_non_root = true
              seccomp_profile {
                type = "RuntimeDefault"
              }
            }

            dynamic "toleration" {
              for_each = try(var.spec.server.scheduling.tolerations, [])
              content {
                key                = try(coalesce(toleration.value.key), "") != "" ? toleration.value.key : null
                operator           = try(coalesce(toleration.value.operator), "") != "" ? toleration.value.operator : null
                value              = try(coalesce(toleration.value.value), "") != "" ? toleration.value.value : null
                effect             = try(coalesce(toleration.value.effect), "") != "" ? toleration.value.effect : null
                toleration_seconds = try(toleration.value.toleration_seconds, null)
              }
            }

            init_container {
              name    = "snapshot"
              image   = local.openbao_image
              command = ["sh", "${local.backup_scripts_mount}/snapshot.sh"]

              dynamic "env" {
                for_each = { for k in sort(keys(local.snapshot_env)) : k => local.snapshot_env[k] }
                content {
                  name  = env.key
                  value = env.value
                }
              }

              volume_mount {
                name       = "scripts"
                mount_path = local.backup_scripts_mount
                read_only  = true
              }
              volume_mount {
                name       = "snapshots"
                mount_path = local.backup_snapshots_mount
              }
              dynamic "volume_mount" {
                for_each = local.tls_enabled ? [1] : []
                content {
                  name       = "tls"
                  mount_path = local.tls_mount_path
                  read_only  = true
                }
              }

              security_context {
                allow_privilege_escalation = false
                capabilities {
                  drop = ["ALL"]
                }
              }

              dynamic "resources" {
                for_each = local.backup_resources != null ? [local.backup_resources] : []
                content {
                  requests = try(resources.value.requests, null)
                  limits   = try(resources.value.limits, null)
                }
              }
            }

            container {
              name    = "upload"
              image   = local.rclone_image
              command = ["sh", "${local.backup_scripts_mount}/upload.sh"]

              dynamic "env" {
                for_each = { for k in sort(keys(local.upload_env)) : k => local.upload_env[k] }
                content {
                  name  = env.key
                  value = env.value
                }
              }
              dynamic "env" {
                for_each = local.backup_has_credentials ? local.backup_credential_env_keys : []
                content {
                  name = env.value
                  value_from {
                    secret_key_ref {
                      name = local.backup_credentials_name
                      key  = env.value
                    }
                  }
                }
              }

              volume_mount {
                name       = "scripts"
                mount_path = local.backup_scripts_mount
                read_only  = true
              }
              volume_mount {
                name       = "snapshots"
                mount_path = local.backup_snapshots_mount
              }
              dynamic "volume_mount" {
                for_each = local.backup_has_store_ca ? [1] : []
                content {
                  name       = "store-ca"
                  mount_path = local.backup_store_ca_mount
                  read_only  = true
                }
              }

              security_context {
                allow_privilege_escalation = false
                capabilities {
                  drop = ["ALL"]
                }
              }

              dynamic "resources" {
                for_each = local.backup_resources != null ? [local.backup_resources] : []
                content {
                  requests = try(resources.value.requests, null)
                  limits   = try(resources.value.limits, null)
                }
              }
            }

            volume {
              name = "scripts"
              config_map {
                name         = local.backup_scripts_name
                default_mode = "0555"
              }
            }
            volume {
              name = "snapshots"
              empty_dir {}
            }
            dynamic "volume" {
              for_each = local.tls_enabled ? [1] : []
              content {
                name = "tls"
                secret {
                  secret_name = local.tls_secret_name
                }
              }
            }
            dynamic "volume" {
              for_each = local.backup_has_store_ca ? [1] : []
              content {
                name = "store-ca"
                secret {
                  secret_name = local.backup_credentials_name
                  items {
                    key  = "ca.pem"
                    path = "ca.pem"
                  }
                }
              }
            }
          }
        }
      }
    }
  }

  depends_on = [
    kubernetes_service_account_v1.backup,
    kubernetes_config_map_v1.backup_scripts,
    kubernetes_secret_v1.backup_credentials,
  ]
}

# The one-shot restore. A restore Job is a RUN, not a state: it fetches
# one snapshot and installs it exactly once, so declarative semantics
# hinge on its NAME — `<name>-restore-<8 hex>` hashing the declaration
# (local.restore_job_name). An unchanged declaration is a no-op on every
# apply; a changed one is a new Job. Two provider defaults are
# deliberately overridden, in both engines: the deploy never WAITS on
# this Job (it waits for the operator's one manual step — the root-token
# Secret after `bao operator init` — and an awaited Job would hang the
# deploy until timeout), and the Job never EXPIRES (a vanished Job would
# be recreated on the next apply and restore AGAIN over live data). No
# active deadline either: the token wait is unbounded by design; the
# bound that matters lives inside restore.sh.
resource "kubernetes_job_v1" "restore" {
  count = local.backup_enabled && local.restore_declared ? 1 : 0

  metadata {
    name      = local.restore_job_name
    namespace = local.namespace
    labels    = local.labels
  }

  wait_for_completion = false

  spec {
    # The preflight in restore.sh may meet a target still unsealing and
    # exit to retry; the install itself is attempted at most this many
    # times, then the Job stays failed and visible with its log.
    backoff_limit = local.restore_job_backoff

    template {
      metadata {
        labels = local.backup_pod_labels
      }
      spec {
        service_account_name = local.backup_name
        restart_policy       = "OnFailure"
        node_selector        = length(try(var.spec.server.scheduling.node_selector, {})) > 0 ? var.spec.server.scheduling.node_selector : null

        security_context {
          run_as_user     = local.backup_job_run_as_user
          run_as_group    = local.backup_job_run_as_group
          fs_group        = local.backup_job_run_as_group
          run_as_non_root = true
          seccomp_profile {
            type = "RuntimeDefault"
          }
        }

        dynamic "toleration" {
          for_each = try(var.spec.server.scheduling.tolerations, [])
          content {
            key                = try(coalesce(toleration.value.key), "") != "" ? toleration.value.key : null
            operator           = try(coalesce(toleration.value.operator), "") != "" ? toleration.value.operator : null
            value              = try(coalesce(toleration.value.value), "") != "" ? toleration.value.value : null
            effect             = try(coalesce(toleration.value.effect), "") != "" ? toleration.value.effect : null
            toleration_seconds = try(toleration.value.toleration_seconds, null)
          }
        }

        init_container {
          name    = "fetch"
          image   = local.rclone_image
          command = ["sh", "${local.backup_scripts_mount}/fetch.sh"]

          dynamic "env" {
            for_each = { for k in sort(keys(local.fetch_env)) : k => local.fetch_env[k] }
            content {
              name  = env.key
              value = env.value
            }
          }
          dynamic "env" {
            for_each = local.backup_has_credentials ? local.backup_credential_env_keys : []
            content {
              name = env.value
              value_from {
                secret_key_ref {
                  name = local.backup_credentials_name
                  key  = env.value
                }
              }
            }
          }

          volume_mount {
            name       = "scripts"
            mount_path = local.backup_scripts_mount
            read_only  = true
          }
          volume_mount {
            name       = "snapshots"
            mount_path = local.backup_snapshots_mount
          }
          dynamic "volume_mount" {
            for_each = local.backup_has_store_ca ? [1] : []
            content {
              name       = "store-ca"
              mount_path = local.backup_store_ca_mount
              read_only  = true
            }
          }

          security_context {
            allow_privilege_escalation = false
            capabilities {
              drop = ["ALL"]
            }
          }

          dynamic "resources" {
            for_each = local.backup_resources != null ? [local.backup_resources] : []
            content {
              requests = try(resources.value.requests, null)
              limits   = try(resources.value.limits, null)
            }
          }
        }

        container {
          name    = "restore"
          image   = local.openbao_image
          command = ["sh", "${local.backup_scripts_mount}/restore.sh"]

          dynamic "env" {
            for_each = { for k in sort(keys(local.restore_env)) : k => local.restore_env[k] }
            content {
              name  = env.key
              value = env.value
            }
          }
          # The operator's one manual step: the main container cannot
          # start until this Secret exists — Kubernetes holds the pod in
          # CreateContainerConfigError, which is the declared shape of
          # "waiting for your one manual step".
          env {
            name = "BAO_TOKEN"
            value_from {
              secret_key_ref {
                name = local.restore.root_token.name
                key  = local.restore.root_token.key
              }
            }
          }

          volume_mount {
            name       = "scripts"
            mount_path = local.backup_scripts_mount
            read_only  = true
          }
          volume_mount {
            name       = "snapshots"
            mount_path = local.backup_snapshots_mount
          }
          dynamic "volume_mount" {
            for_each = local.tls_enabled ? [1] : []
            content {
              name       = "tls"
              mount_path = local.tls_mount_path
              read_only  = true
            }
          }

          security_context {
            allow_privilege_escalation = false
            capabilities {
              drop = ["ALL"]
            }
          }

          dynamic "resources" {
            for_each = local.backup_resources != null ? [local.backup_resources] : []
            content {
              requests = try(resources.value.requests, null)
              limits   = try(resources.value.limits, null)
            }
          }
        }

        volume {
          name = "scripts"
          config_map {
            name         = local.backup_scripts_name
            default_mode = "0555"
          }
        }
        volume {
          name = "snapshots"
          empty_dir {}
        }
        dynamic "volume" {
          for_each = local.tls_enabled ? [1] : []
          content {
            name = "tls"
            secret {
              secret_name = local.tls_secret_name
            }
          }
        }
        dynamic "volume" {
          for_each = local.backup_has_store_ca ? [1] : []
          content {
            name = "store-ca"
            secret {
              secret_name = local.backup_credentials_name
              items {
                key  = "ca.pem"
                path = "ca.pem"
              }
            }
          }
        }
      }
    }
  }

  depends_on = [
    kubernetes_service_account_v1.backup,
    kubernetes_config_map_v1.backup_scripts,
    kubernetes_secret_v1.backup_credentials,
  ]
}
