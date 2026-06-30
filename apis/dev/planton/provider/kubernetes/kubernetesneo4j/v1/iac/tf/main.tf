##############################################
# main.tf
#
# Deploys Neo4j Community Edition on Kubernetes
# using the official Neo4j Helm chart.
##############################################

# Conditionally create namespace for Neo4j deployment
resource "kubernetes_namespace_v1" "neo4j_namespace" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.labels
  }
}

# Deploy Neo4j using Helm chart
resource "helm_release" "neo4j" {
  name       = var.metadata.name
  repository = local.neo4j_helm_chart_repo
  chart      = local.neo4j_helm_chart_name
  version    = local.neo4j_helm_chart_version
  namespace  = local.namespace

  values = [
    yamlencode({
      neo4j = {
        name = var.metadata.name

        # Let the chart create its own secret and password
        # The chart will create a secret named "<release>-auth" with key "neo4j-password"

        # Resource limits
        resources = {
          cpu    = var.spec.container.resources.limits.cpu
          memory = var.spec.container.resources.limits.memory
        }

        # Accept Neo4j Community Edition license
        acceptLicenseAgreement = "yes"
      }

      # External service configuration for ingress
      externalService = {
        enabled = local.ingress_enabled
        type    = local.ingress_enabled ? "LoadBalancer" : ""
        annotations = local.ingress_enabled && local.ingress_external_hostname != "" ? {
          "external-dns.alpha.kubernetes.io/hostname" = local.ingress_external_hostname
        } : {}
      }

      # Persistent storage configuration
      volumes = {
        data = {
          mode = "defaultStorageClass"
          size = var.spec.container.disk_size
        }
      }

      # Neo4j configuration overrides (neo4j.conf)
      config = merge(
        {},
        local.heap_max != "" ? {
          "server.memory.heap.initial_size" = local.heap_max
        } : {},
        local.page_cache != "" ? {
          "server.memory.pagecache.size" = local.page_cache
        } : {}
      )

      # Pod labels
      podLabels = local.labels
    })
  ]

  depends_on = [
    kubernetes_namespace_v1.neo4j_namespace
  ]
}

