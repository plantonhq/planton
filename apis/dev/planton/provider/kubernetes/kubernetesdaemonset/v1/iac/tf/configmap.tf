##############################################
# configmap.tf
#
# Creates ConfigMap resources for the DaemonSet.
# ConfigMaps can be referenced in volume mounts.
##############################################

resource "kubernetes_config_map" "this" {
  for_each = var.spec.config_maps

  metadata {
    name      = each.key
    namespace = local.namespace
    labels    = local.final_labels
  }

  data = {
    "config" = each.value
  }

  depends_on = [kubernetes_namespace.this]
}

