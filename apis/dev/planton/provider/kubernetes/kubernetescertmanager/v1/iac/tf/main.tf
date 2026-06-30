# Conditionally create namespace for cert-manager
resource "kubernetes_namespace" "cert_manager" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name = local.namespace
  }
}

# Look up existing namespace when not creating
data "kubernetes_namespace" "cert_manager" {
  count = var.spec.create_namespace ? 0 : 1

  metadata {
    name = local.namespace
  }
}

# Create ServiceAccount with workload identity annotations
resource "kubernetes_service_account" "cert_manager" {
  metadata {
    name        = local.ksa_name
    namespace   = local.namespace_name
    annotations = local.sa_annotations
  }
}

# Deploy cert-manager Helm chart
resource "helm_release" "cert_manager" {
  name       = local.helm_chart_name
  repository = local.helm_chart_repo
  chart      = local.helm_chart_name
  version    = local.helm_chart_version
  namespace  = local.namespace_name

  wait          = true
  wait_for_jobs = true
  timeout       = 180

  atomic          = true
  cleanup_on_fail = true

  values = [yamlencode({
    installCRDs = true
    serviceAccount = {
      create = false
      name   = local.ksa_name
    }
    extraArgs = [
      "--dns01-recursive-nameservers-only",
      "--dns01-recursive-nameservers=1.1.1.1:53,8.8.8.8:53"
    ]
    image = {
      tag = var.spec.kubernetes_cert_manager_version
    }
    startupapicheck = {
      enabled = !var.spec.skip_install_self_signed_issuer
    }
  })]

  depends_on = [kubernetes_service_account.cert_manager]
}
