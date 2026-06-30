resource "helm_release" "helm_chart" {
  # Set the release name to your resource_id from locals
  name = local.resource_id

  # The Helm repo, chart, and version
  repository = local.helm_repo
  chart      = local.helm_chart
  version    = local.helm_version

  # Use created namespace if available, otherwise use namespace name directly
  namespace = local.create_namespace ? kubernetes_namespace.helm_release_namespace[0].metadata[0].name : local.namespace_name

  # Ensure the namespace is created before the Helm release is installed.
  depends_on = [kubernetes_namespace.helm_release_namespace]

  # # "values" must be a list of YAML strings
  # values = [
  #   yamlencode(local.helm_values)
  # ]
}
