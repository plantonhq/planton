# The flag file as one ConfigMap named after the resource (Pulumi twin:
# module/main.go). A KubernetesGoFeatureFlag relay's `config_map` retriever
# reads it through the Kubernetes API on every poll, so an edit reaches
# evaluations within one polling interval and nothing restarts.
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
