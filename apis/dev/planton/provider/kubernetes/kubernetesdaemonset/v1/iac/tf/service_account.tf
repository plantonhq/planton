##############################################
# service_account.tf
#
# Creates ServiceAccount and RBAC resources for the DaemonSet:
#  - ServiceAccount (optional, when create_service_account is true)
#  - ClusterRole and ClusterRoleBinding (for cluster-wide permissions)
#  - Role and RoleBinding (for namespace-scoped permissions)
##############################################

# ServiceAccount
resource "kubernetes_service_account" "this" {
  count = var.spec.create_service_account ? 1 : 0

  metadata {
    name      = local.service_account_name
    namespace = local.namespace
    labels    = local.final_labels
  }

  depends_on = [kubernetes_namespace.this]
}

# ClusterRole for cluster-wide permissions
resource "kubernetes_cluster_role" "this" {
  count = var.spec.create_service_account && var.spec.rbac != null && length(try(var.spec.rbac.cluster_rules, [])) > 0 ? 1 : 0

  metadata {
    name   = "${var.metadata.name}-cluster-role"
    labels = local.final_labels
  }

  dynamic "rule" {
    for_each = var.spec.rbac.cluster_rules
    content {
      api_groups     = rule.value.api_groups
      resources      = rule.value.resources
      verbs          = rule.value.verbs
      resource_names = try(rule.value.resource_names, null)
    }
  }
}

# ClusterRoleBinding
resource "kubernetes_cluster_role_binding" "this" {
  count = var.spec.create_service_account && var.spec.rbac != null && length(try(var.spec.rbac.cluster_rules, [])) > 0 ? 1 : 0

  metadata {
    name   = "${var.metadata.name}-cluster-role-binding"
    labels = local.final_labels
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = kubernetes_cluster_role.this[0].metadata[0].name
  }

  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account.this[0].metadata[0].name
    namespace = local.namespace
  }
}

# Role for namespace-scoped permissions
resource "kubernetes_role" "this" {
  count = var.spec.create_service_account && var.spec.rbac != null && length(try(var.spec.rbac.namespace_rules, [])) > 0 ? 1 : 0

  metadata {
    name      = "${var.metadata.name}-role"
    namespace = local.namespace
    labels    = local.final_labels
  }

  dynamic "rule" {
    for_each = var.spec.rbac.namespace_rules
    content {
      api_groups     = rule.value.api_groups
      resources      = rule.value.resources
      verbs          = rule.value.verbs
      resource_names = try(rule.value.resource_names, null)
    }
  }

  depends_on = [kubernetes_namespace.this]
}

# RoleBinding
resource "kubernetes_role_binding" "this" {
  count = var.spec.create_service_account && var.spec.rbac != null && length(try(var.spec.rbac.namespace_rules, [])) > 0 ? 1 : 0

  metadata {
    name      = "${var.metadata.name}-role-binding"
    namespace = local.namespace
    labels    = local.final_labels
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = kubernetes_role.this[0].metadata[0].name
  }

  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account.this[0].metadata[0].name
    namespace = local.namespace
  }

  depends_on = [kubernetes_namespace.this]
}

