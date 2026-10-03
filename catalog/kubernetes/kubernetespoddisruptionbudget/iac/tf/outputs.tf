# Outputs — must flatten onto KubernetesPodDisruptionBudgetOutputs
# (outputs.proto) identically to the Pulumi module's exports.

output "pod_disruption_budget_name" {
  description = "The name of the PodDisruptionBudget object as created in the cluster"
  value       = kubernetes_pod_disruption_budget_v1.pod_disruption_budget.metadata[0].name
}

output "namespace" {
  description = "The namespace the budget was created in"
  value       = kubernetes_pod_disruption_budget_v1.pod_disruption_budget.metadata[0].namespace
}
