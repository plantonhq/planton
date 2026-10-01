output "name" {
  description = "Full resource name of the policy (projects/{project}/locations/{location}/deployPolicies/{deploy_policy_id})"
  value       = google_clouddeploy_deploy_policy.this.id
}

output "deploy_policy_id" {
  description = "The policy's ID"
  value       = google_clouddeploy_deploy_policy.this.name
}

output "uid" {
  description = "Google's unique identifier for the policy"
  value       = google_clouddeploy_deploy_policy.this.uid
}
