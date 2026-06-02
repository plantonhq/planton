resource "kubernetes_manifest" "destination_rule" {
  manifest = {
    apiVersion = "networking.istio.io/v1"
    kind       = "DestinationRule"

    metadata = {
      name      = local.destination_rule_name
      namespace = local.namespace
      labels    = local.labels
    }

    spec = local.destination_rule_spec
  }
}
