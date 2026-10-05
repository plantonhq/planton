# Outputs - identical names in the Pulumi module
# (KubernetesGoFeatureFlagFlagFileOutputs).

output "config_map_name" {
  description = "Name of the rendered ConfigMap (= metadata.name)"
  value       = local.name
}

output "key" {
  description = "ConfigMap data key holding the flag file"
  value       = local.key
}

output "namespace" {
  description = "Namespace of the rendered ConfigMap"
  value       = local.namespace
}
