# Conditionally create namespace if create_namespace is true
resource "kubernetes_namespace" "ingress_nginx" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.final_labels
  }
}

resource "helm_release" "ingress_nginx" {
  name       = local.release_name
  namespace  = local.namespace
  repository = local.helm_chart_repo
  chart      = local.helm_chart_name
  version    = local.chart_version

  create_namespace = false
  atomic           = true
  cleanup_on_fail  = true
  wait_for_jobs    = true
  timeout          = 180

  # helm provider v3 replaced `set {}`/`dynamic "set"` blocks with list attributes;
  # use the house values=[yamlencode(...)] idiom, which also lets annotation keys
  # (containing "/" and ".") be set literally instead of via --set escaping.
  # Mirrors the Pulumi module's controller values.
  values = [
    yamlencode({
      controller = {
        service = {
          type        = local.service_type
          annotations = local.service_annotations
        }
        ingressClassResource = {
          default = true
        }
        watchIngressWithoutClass = true
      }
    })
  ]
}

