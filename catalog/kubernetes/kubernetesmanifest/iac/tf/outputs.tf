# Outputs — must flatten onto KubernetesManifestOutputs
# (outputs.proto) identically to the Pulumi module's exports.

output "namespace" {
  description = "The anchor namespace: where namespaced documents without an explicit metadata.namespace were applied"
  value       = local.namespace_name
}

output "applied_resources" {
  description = "The applied-resource inventory, one apiVersion/Kind/name entry per document in manifest order (derived from the input YAML, identical on both engines)"
  value       = local.applied_resources
}
