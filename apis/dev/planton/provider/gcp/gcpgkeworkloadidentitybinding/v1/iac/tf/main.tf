# This Terraform module implements GKE Workload Identity binding by granting
# "roles/iam.workloadIdentityUser" on a Google Service Account (GSA) to the
# specified Kubernetes Service Account (KSA). This enables pods using the KSA
# to impersonate the GSA when accessing Google Cloud APIs.

locals {
  # Construct the Workload Identity member string from the KSA details.
  # Format: serviceAccount:<project-id>.svc.id.goog[<namespace>/<ksa-name>]
  workload_identity_member = "serviceAccount:${var.project_id}.svc.id.goog[${var.ksa_namespace}/${var.ksa_name}]"
}

# Grant the Workload Identity User role to the Kubernetes Service Account
resource "google_service_account_iam_member" "workload_identity_binding" {
  service_account_id = var.service_account_email
  role               = "roles/iam.workloadIdentityUser"
  member             = local.workload_identity_member
}


