# Outputs — must flatten onto KubernetesStorageClassOutputs
# (outputs.proto) identically to the Pulumi module's exports.

output "storage_class_name" {
  description = "The name of the StorageClass object as created in the cluster"
  value       = kubernetes_storage_class_v1.storage_class.metadata[0].name
}

output "provisioner" {
  description = "The provisioner (CSI driver) backing this class"
  value       = kubernetes_storage_class_v1.storage_class.storage_provisioner
}

output "is_default_class" {
  description = "Whether this class is annotated as the cluster's default StorageClass"
  value       = var.spec.is_default_class
}
