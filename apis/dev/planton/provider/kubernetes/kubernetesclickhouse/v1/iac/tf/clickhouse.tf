resource "kubernetes_manifest" "clickhouse_installation" {
  manifest = {
    apiVersion = "clickhouse.altinity.com/v1"
    kind       = "ClickHouseInstallation"

    metadata = {
      name      = local.cluster_name
      namespace = local.namespace
      labels    = local.final_labels
    }

    spec = {
      configuration = {
        # User configuration with password from secret
        users = {
          "${local.default_username}/password_sha256_hex" = {
            k8s_secret = {
              name = kubernetes_secret_v1.clickhouse_password.metadata[0].name
              key  = local.password_secret_key
            }
          }
        }

        # Cluster configuration
        clusters = [
          {
            name = local.cluster_name
            layout = {
              shardsCount   = local.shard_count
              replicasCount = local.replica_count
            }
          }
        ]

        # ZooKeeper configuration (for clustered deployments)
        zookeeper = local.cluster_is_enabled ? local.zookeeper_config : null
      }

      # Default templates
      defaults = {
        templates = {
          podTemplate             = "clickhouse-pod"
          dataVolumeClaimTemplate = "data-volume"
        }
      }

      # Templates
      templates = {
        # Pod template with container resources
        podTemplates = [
          {
            name = "clickhouse-pod"
            spec = {
              containers = [
                {
                  name  = "clickhouse"
                  image = "clickhouse/clickhouse-server:${local.clickhouse_version}"
                  resources = {
                    requests = {
                      cpu    = var.spec.container.resources.requests.cpu
                      memory = var.spec.container.resources.requests.memory
                    }
                    limits = {
                      cpu    = var.spec.container.resources.limits.cpu
                      memory = var.spec.container.resources.limits.memory
                    }
                  }
                }
              ]
            }
          }
        ]

        # Volume claim template for persistence
        volumeClaimTemplates = [
          {
            name = "data-volume"
            spec = {
              accessModes = ["ReadWriteOnce"]
              resources = {
                requests = {
                  storage = var.spec.container.disk_size
                }
              }
            }
          }
        ]
      }
    }
  }

  depends_on = [
    kubernetes_namespace_v1.clickhouse_namespace,
    kubernetes_secret_v1.clickhouse_password
  ]
}
