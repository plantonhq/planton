resource "kubernetes_namespace_v1" "controller" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = var.spec.namespace
    labels = local.labels
  }
}

resource "helm_release" "controller" {
  name             = local.release_name
  namespace        = var.spec.namespace
  create_namespace = false

  chart   = local.chart_oci
  version = var.spec.helm_chart_version

  values = local.helm_values_list

  depends_on = [kubernetes_namespace_v1.controller]
}
