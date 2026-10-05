# `get` on exactly the ConfigMaps the `config_map` retrievers read - one
# Role and RoleBinding, both `<metadata.name>-flag-reader`, in each
# namespace those ConfigMaps live in (Pulumi twin: rbac.go). The relay's
# retriever issues a plain GET per poll; the chart ships no RBAC of its own.

resource "kubernetes_role_v1" "flag_reader" {
  for_each = local.config_map_grants

  metadata {
    name      = local.flag_reader_name
    namespace = each.key
    labels    = local.labels
  }

  rule {
    api_groups     = [""]
    resources      = ["configmaps"]
    resource_names = each.value
    verbs          = ["get"]
  }

  depends_on = [kubernetes_namespace_v1.relay]
}

resource "kubernetes_role_binding_v1" "flag_reader" {
  for_each = local.config_map_grants

  metadata {
    name      = local.flag_reader_name
    namespace = each.key
    labels    = local.labels
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = local.flag_reader_name
  }

  subject {
    kind      = "ServiceAccount"
    name      = local.service_account_name
    namespace = local.namespace
  }

  depends_on = [kubernetes_role_v1.flag_reader]
}
