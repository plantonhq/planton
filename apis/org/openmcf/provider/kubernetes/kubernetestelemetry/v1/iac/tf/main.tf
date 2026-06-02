resource "kubernetes_manifest" "telemetry" {
  manifest = {
    apiVersion = "telemetry.istio.io/v1"
    kind       = "Telemetry"

    metadata = {
      name      = local.telemetry_name
      namespace = local.namespace
      labels    = local.labels
    }

    spec = local.telemetry_spec
  }
}
