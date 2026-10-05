# flagd's SourceConfig array (it carries authorization headers), read into
# FLAGD_SOURCES (Pulumi twin: identity.go sourcesSecret).
resource "kubernetes_secret_v1" "sources" {
  metadata {
    name      = "${local.name}-sources"
    namespace = local.namespace
    labels    = local.labels
  }
  type = "Opaque"
  data = {
    sources = local.sources_json
  }
  depends_on = [kubernetes_namespace_v1.flagd]
}

# flagd's ServiceAccount, unless an existing one is named.
resource "kubernetes_service_account_v1" "flagd" {
  count = local.create_service_account ? 1 : 0
  metadata {
    name        = local.name
    namespace   = local.namespace
    labels      = local.labels
    annotations = try(var.spec.service_account.annotations, {})
  }
  depends_on = [kubernetes_namespace_v1.flagd]
}

# get/list/watch on FeatureFlag resources in each namespace a `feature_flag`
# source reads (flagd watches them through an informer).
resource "kubernetes_role_v1" "flag_reader" {
  for_each = local.feature_flag_namespaces
  metadata {
    name      = "${local.name}-flag-reader"
    namespace = each.value
    labels    = local.labels
  }
  rule {
    api_groups = ["core.openfeature.dev"]
    resources  = ["featureflags"]
    verbs      = ["get", "list", "watch"]
  }
  depends_on = [kubernetes_namespace_v1.flagd]
}

resource "kubernetes_role_binding_v1" "flag_reader" {
  for_each = local.feature_flag_namespaces
  metadata {
    name      = "${local.name}-flag-reader"
    namespace = each.value
    labels    = local.labels
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = "${local.name}-flag-reader"
  }
  subject {
    kind      = "ServiceAccount"
    name      = local.service_account_name
    namespace = local.namespace
  }
  depends_on = [kubernetes_role_v1.flag_reader]
}
