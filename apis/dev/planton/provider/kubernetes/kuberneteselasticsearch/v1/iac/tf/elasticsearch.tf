resource "kubernetes_manifest" "elasticsearch" {
  manifest = yamldecode(
    yamlencode({
      apiVersion = "elasticsearch.k8s.elastic.co/v1"
      kind       = "Elasticsearch"
      metadata = {
        name      = var.metadata.name
        namespace = local.namespace_name
        labels    = local.final_labels
        annotations = {
          "pulumi.com/patchForce" = "true"
        }
      }
      spec = {
        version = local.elasticsearch_version

        # We define a single nodeSet in an array (or more if needed).
        nodeSets = [
          {
            name  = "elasticsearch"
            count = var.spec.elasticsearch.container.replicas

            # Additional configuration
            config = {
              "node.roles"            = ["master", "data", "ingest"]
              "node.store.allow_mmap" = false
            }

            # Pod template, with resource requests/limits
            podTemplate = {
              metadata = {
                labels = {
                  "role" = "master"
                }
              }
              spec = {
                containers = [
                  {
                    name = "elasticsearch"
                    resources = {
                      requests = {
                        memory = var.spec.elasticsearch.container.resources.requests.memory
                        cpu    = var.spec.elasticsearch.container.resources.requests.cpu
                      }
                      limits = {
                        memory = var.spec.elasticsearch.container.resources.limits.memory
                        cpu    = var.spec.elasticsearch.container.resources.limits.cpu
                      }
                    }
                  }
                ]
              }
            }

            # Volume claim templates if persistence is enabled
            volumeClaimTemplates = var.spec.elasticsearch.container.persistence_enabled ? [
              {
                metadata = {
                  name = "elasticsearch-data"
                }
                spec = {
                  accessModes = ["ReadWriteOnce"]
                  resources = {
                    requests = {
                      storage = var.spec.elasticsearch.container.disk_size
                    }
                  }
                }
              }
            ] : []
          }
        ]

        # Disable the ECK self-signed certificate on HTTP
        http = {
          tls = {
            selfSignedCertificate = {
              disabled = true
            }
          }
        }
      }
    })
  )
}

# 2) Optionally create Kibana CR if enabled == true
resource "kubernetes_manifest" "kibana" {
  count = var.spec.kibana.enabled ? 1 : 0

  manifest = yamldecode(
    yamlencode({
      apiVersion = "kibana.k8s.elastic.co/v1beta1"
      kind       = "Kibana"
      metadata = {
        # Same name as ES or adapt as needed
        name      = var.metadata.name
        namespace = local.namespace_name
        labels    = local.final_labels
        annotations = {
          "pulumi.com/patchForce" = "true"
        }
      }
      spec = {
        version = local.elasticsearch_version
        count   = var.spec.kibana.container.replicas

        podTemplate = {
          spec = {
            containers = [
              {
                name = "kibana"
                resources = {
                  requests = {
                    memory = var.spec.kibana.container.resources.requests.memory
                    cpu    = var.spec.kibana.container.resources.requests.cpu
                  }
                  limits = {
                    memory = var.spec.kibana.container.resources.limits.memory
                    cpu    = var.spec.kibana.container.resources.limits.cpu
                  }
                }
              }
            ]
          }
        }

        # Reference our just-created Elasticsearch
        elasticsearchRef = {
          name      = kubernetes_manifest.elasticsearch.manifest["metadata"]["name"]
          namespace = local.namespace_name
        }

        # Disable the self-signed cert for Kibana’s HTTP, same as ES
        http = {
          tls = {
            selfSignedCertificate = {
              disabled = true
            }
          }
        }
      }
    })
  )

  # Ensure Kibana is created after Elasticsearch
  depends_on = [
    kubernetes_manifest.elasticsearch
  ]
}
