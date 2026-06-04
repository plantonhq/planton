# Deploy OpenBao using the official Helm chart
resource "helm_release" "openbao" {
  name       = var.metadata.name
  namespace  = local.namespace
  repository = local.helm_chart_repo
  chart      = local.helm_chart_name
  version    = local.helm_chart_version

  # Wait for namespace to be created if create_namespace is true
  depends_on = [kubernetes_namespace.openbao_namespace]

  # helm provider v3 replaced `set {}`/`dynamic "set"` blocks with list attributes;
  # use the house values=[yamlencode(...)] idiom. local.helm_values mirrors the Pulumi
  # module's values map (incl. native bool/number types instead of tostring()-ed --set
  # strings, and the literal "iam.gke.io/gcp-service-account" annotation key).
  values = [
    yamlencode(local.helm_values)
  ]

}
