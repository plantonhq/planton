# The flag definitions as one ConfigMap named after the resource (Pulumi twin:
# module/main.go). A KubernetesFlagd `config_map` source mounts it as a
# directory, so an edit reaches flagd once the kubelet syncs the volume and
# nothing restarts.
resource "kubernetes_config_map_v1" "flag_file" {
  metadata {
    name      = local.name
    namespace = local.namespace
    labels    = local.labels
  }

  data = {
    (local.key) = local.flag_file
  }
}
