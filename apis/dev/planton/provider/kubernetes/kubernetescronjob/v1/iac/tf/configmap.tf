##############################################
# configmap.tf
#
# ConfigMap resources for the KubernetesCronJob.
#
# This file creates ConfigMaps from the spec.config_maps map.
# Each ConfigMap stores its content under a key with the same
# name as the ConfigMap, matching the Pulumi module behavior.
#
# These ConfigMaps can be mounted as files into containers
# using the spec.volume_mounts configuration.
##############################################

resource "kubernetes_config_map" "this" {
  for_each = try(var.spec.config_maps, {})

  metadata {
    name      = each.key
    namespace = local.namespace_name
    labels    = local.final_labels
  }

  data = {
    (each.key) = each.value
  }

  depends_on = [
    kubernetes_namespace.this
  ]
}

