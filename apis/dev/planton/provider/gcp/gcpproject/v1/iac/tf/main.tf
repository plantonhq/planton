# main.tf - GCP Project resource definitions

# Generate a random 3-character suffix for project_id uniqueness
# Only created when add_suffix is true
resource "random_string" "project_suffix" {
  count = var.spec.add_suffix ? 1 : 0

  length  = 3
  special = false
  upper   = false
  lower   = true
  numeric = false
}

# Create the GCP Project
resource "google_project" "this" {
  name                = var.metadata.name
  project_id          = local.project_id
  billing_account     = local.billing_account_id
  org_id              = local.parent_org_id
  folder_id           = local.parent_folder_id
  labels              = local.gcp_labels
  auto_create_network = local.auto_create_network
  deletion_policy     = local.deletion_policy
}

# Enable specified Google Cloud APIs
resource "google_project_service" "enabled_apis" {
  for_each = toset(local.enabled_apis)

  project = google_project.this.project_id
  service = each.value

  # Don't disable dependent services when this resource is destroyed
  disable_dependent_services = true

  # Don't disable the service when this resource is destroyed
  disable_on_destroy = false

  depends_on = [google_project.this]
}

# Optionally grant Owner role to specified IAM member
resource "google_project_iam_member" "owner" {
  count = local.owner_member != null ? 1 : 0

  project = google_project.this.project_id
  role    = "roles/owner"
  member  = local.owner_member

  depends_on = [google_project.this]
}

