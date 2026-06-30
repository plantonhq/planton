resource "kubernetes_manifest" "solr_cloud" {
  # The 'kubernetes_manifest' resource requires a valid Kubernetes manifest in YAML or JSON form.
  # We can build that as a Terraform map/object and encode it to YAML.

  manifest = yamldecode(
    yamlencode({
      apiVersion = "solr.apache.org/v1beta1"
      kind       = "SolrCloud"

      metadata = {
        name      = var.metadata.name
        namespace = local.namespace
        labels    = local.final_labels
      }

      spec = {
        replicas = var.spec.solr_container.replicas

        solrImage = {
          repository = var.spec.solr_container.image.repo
          tag        = var.spec.solr_container.image.tag
        }

        solrJavaMem = try(var.spec.config.java_mem, "")
        solrOpts    = try(var.spec.config.opts, "")
        solrGCTune  = try(var.spec.config.garbage_collection_tuning, "")

        solrModules = [
          "jaegertracer-configurator",
          "ltr",
        ]

        customSolrKubeOptions = {
          podOptions = {
            resources = {
              limits = {
                # Include CPU if desired
                "memory" = var.spec.solr_container.resources.limits.memory
              }
              requests = {
                "cpu"    = var.spec.solr_container.resources.requests.cpu
                "memory" = var.spec.solr_container.resources.requests.memory
              }
            }
          }
        }

        dataStorage = {
          ephemeral = null
          persistent = {
            reclaimPolicy = "Delete"
            pvcTemplate = {
              spec = {
                resources = {
                  requests = {
                    "storage" = var.spec.solr_container.disk_size
                  }
                }
              }
            }
          }
        }

        zookeeperRef = {
          provided = {
            replicas = var.spec.zookeeper_container.replicas

            persistence = {
              spec = {
                resources = {
                  requests = {
                    "storage" = var.spec.zookeeper_container.disk_size
                  }
                }
              }
            }

            zookeeperPodPolicy = {
              resources = {
                limits = {
                  # Include CPU if desired
                  "memory" = var.spec.zookeeper_container.resources.limits.memory
                }
                requests = {
                  "cpu"    = var.spec.zookeeper_container.resources.requests.cpu
                  "memory" = var.spec.zookeeper_container.resources.requests.memory
                }
              }
            }
          }
        }
      }
    })
  )
}
