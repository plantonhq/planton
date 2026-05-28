# Create the cluster-scoped Gateway API GatewayClass custom resource.
resource "kubernetes_manifest" "gateway_class" {
  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "GatewayClass"

    metadata = {
      name   = local.gateway_class_name
      labels = local.labels
    }

    spec = merge(
      {
        controllerName = local.controller_name
      },
      var.spec.description != null ? { description = var.spec.description } : {},
      local.parameters_ref != null ? { parametersRef = local.parameters_ref } : {}
    )
  }
}
