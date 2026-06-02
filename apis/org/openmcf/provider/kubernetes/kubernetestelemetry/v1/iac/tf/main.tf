resource "kubernetes_manifest" "telemetry" {
  manifest = {
    apiVersion = "telemetry.istio.io/v1"
    kind       = "Telemetry"

    metadata = {
      name      = var.metadata.name
      namespace = var.spec.namespace
      labels    = local.labels
    }

    spec = local.manifest_spec
  }
}
