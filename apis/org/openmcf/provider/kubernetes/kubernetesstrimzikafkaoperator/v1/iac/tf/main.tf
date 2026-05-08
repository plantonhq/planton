resource "kubernetes_namespace" "operator_namespace" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.labels
  }
}

resource "helm_release" "strimzi_kafka_operator" {
  name             = local.helm_release_name
  namespace        = local.namespace
  repository       = local.helm_chart_repo
  chart            = local.helm_chart_name
  version          = local.helm_chart_version
  create_namespace = false
  atomic           = false
  cleanup_on_fail  = true
  wait_for_jobs    = true
  timeout          = 180

  values = [
    yamlencode({
      watchAnyNamespace = true
    })
  ]

  depends_on = [kubernetes_namespace.operator_namespace]
}
