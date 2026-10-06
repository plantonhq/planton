# Agent teammates' read-only way into Grafana (spec.agent_reader). Twin of
# the Pulumi module's agentreader.go.
#
# Grafana cannot provision service accounts from files, so a Job run after
# the release is Ready keeps a Viewer account and its one token: the
# ServiceAccount the Job runs as, a Role and RoleBinding granting it
# exactly the token Secret, the scripts ConfigMap, and the Job.
#
# The token Secret is NOT a resource here. The Job writes it with an owner
# reference to the ServiceAccount below: Kubernetes deletes it when the
# ServiceAccount goes (the block removed, or the resource destroyed), and
# the token never passes through state, where a module-owned Secret's data
# would land on every refresh.

resource "kubernetes_service_account_v1" "agent_reader" {
  count = local.agent_reader_declared ? 1 : 0

  metadata {
    name      = local.agent_reader_name
    namespace = local.namespace
    labels    = local.labels
  }

  depends_on = [helm_release.grafana]
}

# Kubernetes cannot scope `create` to a name (the name is not known before
# the object exists), so create is namespace-wide; everything else names
# the one Secret, and `get` names the one ServiceAccount whose uid the
# owner reference carries.
resource "kubernetes_role_v1" "agent_reader" {
  count = local.agent_reader_declared ? 1 : 0

  metadata {
    name      = local.agent_reader_name
    namespace = local.namespace
    labels    = local.labels
  }

  rule {
    api_groups     = [""]
    resources      = ["secrets"]
    resource_names = [local.agent_reader_name]
    verbs          = ["get", "update", "delete"]
  }

  rule {
    api_groups = [""]
    resources  = ["secrets"]
    verbs      = ["create"]
  }

  rule {
    api_groups     = [""]
    resources      = ["serviceaccounts"]
    resource_names = [local.agent_reader_name]
    verbs          = ["get"]
  }

  depends_on = [helm_release.grafana]
}

resource "kubernetes_role_binding_v1" "agent_reader" {
  count = local.agent_reader_declared ? 1 : 0

  metadata {
    name      = local.agent_reader_name
    namespace = local.namespace
    labels    = local.labels
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = local.agent_reader_name
  }

  subject {
    kind      = "ServiceAccount"
    name      = local.agent_reader_name
    namespace = local.namespace
  }

  depends_on = [kubernetes_role_v1.agent_reader]
}

resource "kubernetes_config_map_v1" "agent_reader_script" {
  count = local.agent_reader_declared ? 1 : 0

  metadata {
    name      = local.agent_reader_script_name
    namespace = local.namespace
    labels    = local.labels
  }

  data = {
    "agent-reader.sh" = local.agent_reader_script
  }

  depends_on = [helm_release.grafana]
}

# The Job is a RUN, so its name is its identity (local.agent_reader_job_name
# hashes what it reconciles). The apply waits for it: it depends on nothing
# but a Ready Grafana, so a token that fails to mint fails this apply
# rather than an agent's first call. No ttl_seconds_after_finished: a
# finished Job that vanished would come back on the next apply as drift,
# and the Job object is where its log stays readable. The script bounds its
# own wait for Grafana; backoff and the deadline bound the run, retries
# included.
resource "kubernetes_job_v1" "agent_reader" {
  count = local.agent_reader_declared ? 1 : 0

  metadata {
    name      = local.agent_reader_job_name
    namespace = local.namespace
    labels    = local.labels
  }

  wait_for_completion = true

  timeouts {
    create = "15m"
    update = "15m"
  }

  spec {
    backoff_limit           = 6
    active_deadline_seconds = 900

    template {
      metadata {
        labels = local.labels
      }

      spec {
        service_account_name = local.agent_reader_name
        restart_policy       = "OnFailure"
        node_selector        = length(try(var.spec.scheduling.node_selector, {})) > 0 ? var.spec.scheduling.node_selector : null
        priority_class_name  = try(coalesce(var.spec.scheduling.priority_class_name), null)

        # The Job runs as `nobody` on a read-only root filesystem; kubectl
        # writes its cache under HOME, an emptyDir.
        security_context {
          run_as_user     = 65534
          run_as_group    = 65534
          run_as_non_root = true
          seccomp_profile {
            type = "RuntimeDefault"
          }
        }

        dynamic "toleration" {
          for_each = try(var.spec.scheduling.tolerations, [])
          content {
            key                = try(coalesce(toleration.value.key), "") != "" ? toleration.value.key : null
            operator           = try(coalesce(toleration.value.operator), "") != "" ? toleration.value.operator : null
            value              = try(coalesce(toleration.value.value), "") != "" ? toleration.value.value : null
            effect             = try(coalesce(toleration.value.effect), "") != "" ? toleration.value.effect : null
            toleration_seconds = try(toleration.value.toleration_seconds, null)
          }
        }

        dynamic "image_pull_secrets" {
          for_each = local.agent_reader_pull_secret_name != "" ? [local.agent_reader_pull_secret_name] : []
          content {
            name = image_pull_secrets.value
          }
        }

        container {
          name    = "agent-reader"
          image   = local.agent_reader_image
          command = ["sh", "/scripts/agent-reader.sh"]

          dynamic "env" {
            for_each = local.agent_reader_env
            content {
              name  = env.value[0]
              value = env.value[1]
            }
          }

          env {
            name = "GRAFANA_ADMIN_USER"
            value_from {
              secret_key_ref {
                name = local.admin_secret_name
                key  = local.agent_reader_admin_user_key
              }
            }
          }

          env {
            name = "GRAFANA_ADMIN_PASSWORD"
            value_from {
              secret_key_ref {
                name = local.admin_secret_name
                key  = local.agent_reader_admin_password_key
              }
            }
          }

          resources {
            requests = {
              cpu    = "50m"
              memory = "64Mi"
            }
            limits = {
              cpu    = "200m"
              memory = "128Mi"
            }
          }

          security_context {
            allow_privilege_escalation = false
            read_only_root_filesystem  = true
            capabilities {
              drop = ["ALL"]
            }
          }

          volume_mount {
            name       = "scripts"
            mount_path = "/scripts"
          }

          volume_mount {
            name       = "home"
            mount_path = "/tmp"
          }
        }

        volume {
          name = "scripts"
          config_map {
            name         = local.agent_reader_script_name
            default_mode = "0555"
          }
        }

        volume {
          name = "home"
          empty_dir {}
        }
      }
    }
  }

  depends_on = [
    helm_release.grafana,
    kubernetes_service_account_v1.agent_reader,
    kubernetes_role_binding_v1.agent_reader,
    kubernetes_config_map_v1.agent_reader_script,
  ]
}
