# Outputs - identical names in the Pulumi module
# (KubernetesFlagdFlagFileOutputs).

output "config_map_name" {
  description = "Name of the rendered ConfigMap (= metadata.name)"
  value       = local.name
}

output "key" {
  description = "ConfigMap data key holding the flag definitions"
  value       = local.key
}

output "namespace" {
  description = "Namespace of the rendered ConfigMap"
  value       = local.namespace
}
