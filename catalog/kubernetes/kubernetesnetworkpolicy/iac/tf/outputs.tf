# Outputs — must flatten onto KubernetesNetworkPolicyOutputs
# (outputs.proto) identically to the Pulumi module's exports.

output "network_policy_name" {
  description = "The name of the NetworkPolicy object as created in the cluster"
  value       = kubernetes_network_policy_v1.network_policy.metadata[0].name
}

output "namespace" {
  description = "The namespace the NetworkPolicy was created in"
  value       = kubernetes_network_policy_v1.network_policy.metadata[0].namespace
}

output "policy_types" {
  description = "The governed directions as deployed, including inferred types"
  value       = local.policy_types_string
}
