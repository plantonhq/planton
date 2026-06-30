# Outputs mirror the GcpGkeWorkloadIdentityBindingStackOutputs proto message

output "member" {
  description = "The IAM member string added to the policy, e.g. 'serviceAccount:my-project.svc.id.goog[cert-manager/cert-manager]'"
  value       = google_service_account_iam_member.workload_identity_binding.member
}

output "service_account_email" {
  description = "The bound GSA email (echoed from input for convenience)"
  value       = var.service_account_email
}


