##############################################
# configmap.tf
#
# ConfigMap resources for the KubernetesDeployment.
#
# This file creates ConfigMaps from the spec.config_maps map.
# ConfigMap names are prefixed with metadata.name to avoid conflicts
# when multiple deployments share a namespace.
# Each ConfigMap stores its content under a key with the original
# config_maps key name, matching the Pulumi module behavior.
#
# These ConfigMaps can be mounted as files into containers
# using the spec.container.app.volume_mounts configuration.
##############################################

resource "kubernetes_config_map" "this" {
  for_each = try(var.spec.config_maps, {})

  metadata {
    # Prefix ConfigMap name with metadata.name to avoid conflicts when multiple deployments share a namespace
    name      = "${var.metadata.name}-${each.key}"
    namespace = local.namespace
    labels    = local.final_labels
  }

  data = {
    (each.key) = each.value
  }

  depends_on = [
    kubernetes_namespace.this
  ]
}

