# Create the namespaced Gateway API Gateway custom resource.
resource "kubernetes_manifest" "gateway" {
  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "Gateway"

    metadata = {
      name      = local.gateway_name
      namespace = local.namespace
      labels    = local.labels
    }

    spec = local.gateway_spec
  }
}
