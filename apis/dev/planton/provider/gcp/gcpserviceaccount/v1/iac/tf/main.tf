# Create the GCP Service Account
resource "google_service_account" "main" {
  account_id   = var.spec.service_account_id
  display_name = var.metadata.name
  project      = var.spec.project_id.value
  description  = "Service account managed by Planton for ${var.metadata.name}"
}

# Optionally create a JSON key for the service account
# Only created if spec.create_key is explicitly set to true
resource "google_service_account_key" "main" {
  count = local.create_key ? 1 : 0

  service_account_id = google_service_account.main.name

  # Public key type is not needed for private key creation
  # The private key will be base64-encoded JSON
}

# Grant project-level IAM roles to the service account
resource "google_project_iam_member" "project_roles" {
  for_each = toset(local.project_iam_roles)

  project = var.spec.project_id.value
  role    = each.value
  member  = "serviceAccount:${local.service_account_email}"

  depends_on = [google_service_account.main]
}

# Grant organization-level IAM roles to the service account
resource "google_organization_iam_member" "org_roles" {
  for_each = toset(local.org_iam_roles)

  org_id = var.spec.org_id
  role   = each.value
  member = "serviceAccount:${local.service_account_email}"

  depends_on = [google_service_account.main]
}

