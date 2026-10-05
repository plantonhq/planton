# Outputs — must flatten onto KubernetesPersistentVolumeClaimOutputs
# (outputs.proto) identically to the Pulumi module's exports.

output "pvc_name" {
  description = "The name of the PersistentVolumeClaim object as created in the cluster"
  value       = kubernetes_persistent_volume_claim_v1.persistent_volume_claim.metadata[0].name
}

output "namespace" {
  description = "The namespace the claim was created in"
  value       = kubernetes_persistent_volume_claim_v1.persistent_volume_claim.metadata[0].namespace
}

output "storage_request" {
  description = "The requested storage size as a Kubernetes quantity"
  value       = var.spec.storage_request
}
