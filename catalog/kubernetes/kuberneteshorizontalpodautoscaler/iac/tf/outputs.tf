# Outputs — must flatten onto
# KubernetesHorizontalPodAutoscalerOutputs (outputs.proto)
# identically to the Pulumi module's exports.

output "horizontal_pod_autoscaler_name" {
  description = "The name of the HorizontalPodAutoscaler object as created in the cluster"
  value       = kubernetes_horizontal_pod_autoscaler_v2.horizontal_pod_autoscaler.metadata[0].name
}

output "namespace" {
  description = "The namespace the autoscaler was created in"
  value       = kubernetes_horizontal_pod_autoscaler_v2.horizontal_pod_autoscaler.metadata[0].namespace
}

output "scale_target" {
  description = "The scale target as Kind/name"
  value       = local.scale_target_string
}

output "min_replicas" {
  description = "The configured replica floor"
  value       = local.min_replicas
}

output "max_replicas" {
  description = "The configured replica ceiling"
  value       = var.spec.max_replicas
}
