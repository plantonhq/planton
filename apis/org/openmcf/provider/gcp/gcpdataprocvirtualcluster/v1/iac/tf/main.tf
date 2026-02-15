###############################################################################
# Dataproc Virtual Cluster (Dataproc on GKE)
###############################################################################

resource "google_dataproc_cluster" "virtual_cluster" {
  name    = local.cluster_name
  region  = var.spec.region
  project = var.spec.project_id.value
  labels  = local.labels

  virtual_cluster_config {
    dynamic "auxiliary_services_config" {
      for_each = var.spec.auxiliary_services_config != null ? [var.spec.auxiliary_services_config] : []
      content {
        dynamic "metastore_config" {
          for_each = auxiliary_services_config.value.metastore_service != "" ? [auxiliary_services_config.value.metastore_service] : []
          content {
            dataproc_metastore_service = metastore_config.value
          }
        }
        dynamic "spark_history_server_config" {
          for_each = auxiliary_services_config.value.spark_history_server_cluster != "" ? [auxiliary_services_config.value.spark_history_server_cluster] : []
          content {
            dataproc_cluster = spark_history_server_config.value
          }
        }
      }
    }

    kubernetes_cluster_config {
      kubernetes_namespace = var.spec.kubernetes_namespace != null ? var.spec.kubernetes_namespace.value : null

      kubernetes_software_config {
        component_version = var.spec.software_config.component_version
        properties        = var.spec.software_config.properties
      }

      gke_cluster_config {
        gke_cluster_target = var.spec.gke_cluster_target.value

        dynamic "node_pool_target" {
          for_each = var.spec.node_pool_targets
          content {
            node_pool = node_pool_target.value.node_pool.value
            roles     = node_pool_target.value.roles

            dynamic "node_pool_config" {
              for_each = node_pool_target.value.node_pool_config != null ? [node_pool_target.value.node_pool_config] : []
              content {
                locations = node_pool_config.value.locations

                dynamic "autoscaling" {
                  for_each = node_pool_config.value.autoscaling != null ? [node_pool_config.value.autoscaling] : []
                  content {
                    min_node_count = autoscaling.value.min_node_count
                    max_node_count = autoscaling.value.max_node_count
                  }
                }

                dynamic "config" {
                  for_each = (
                    node_pool_config.value.machine_type != "" ||
                    node_pool_config.value.local_ssd_count > 0 ||
                    node_pool_config.value.min_cpu_platform != "" ||
                    node_pool_config.value.preemptible ||
                    node_pool_config.value.spot
                  ) ? [1] : []
                  content {
                    machine_type     = node_pool_config.value.machine_type != "" ? node_pool_config.value.machine_type : null
                    local_ssd_count  = node_pool_config.value.local_ssd_count > 0 ? node_pool_config.value.local_ssd_count : null
                    min_cpu_platform = node_pool_config.value.min_cpu_platform != "" ? node_pool_config.value.min_cpu_platform : null
                    preemptible      = node_pool_config.value.preemptible ? true : null
                    spot             = node_pool_config.value.spot ? true : null
                  }
                }
              }
            }
          }
        }
      }
    }

    staging_bucket = var.spec.staging_bucket != null ? var.spec.staging_bucket.value : null
  }
}
