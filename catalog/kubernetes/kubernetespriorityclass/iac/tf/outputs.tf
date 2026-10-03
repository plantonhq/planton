# Outputs — must flatten onto KubernetesPriorityClassOutputs
# (outputs.proto) identically to the Pulumi module's exports.

output "priority_class_name" {
  description = "The name of the PriorityClass object as created in the cluster"
  value       = kubernetes_priority_class_v1.priority_class.metadata[0].name
}

output "value" {
  description = "The priority integer pods of this class receive"
  value       = kubernetes_priority_class_v1.priority_class.value
}
