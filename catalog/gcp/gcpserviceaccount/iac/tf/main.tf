# The service account identity. account_id and project are immutable (ForceNew
# in the provider): changing either destroys and recreates the account, which
# invalidates every IAM binding and workload identity referencing the old email.
# display_name, description, and disabled are all mutable in place.
resource "google_service_account" "main" {
  account_id   = var.spec.service_account_id
  display_name = local.display_name
  project      = local.project_id
  description  = var.spec.description

  # GCP flips disabled state via separate Enable/Disable API calls (not the
  # regular update mask); the provider handles that internally, so toggling
  # this field never recreates the account.
  disabled = local.disabled

  # Adopt an existing account with the same email instead of failing —
  # idempotent bootstrap flows that may race other provisioning paths.
  # Sent only when enabled, the same posture as the Pulumi module: the flag
  # is provider-side (never sent to the API), so an explicit false would only
  # add a null -> false diff on every existing deployment.
  create_ignore_already_exists = var.spec.create_ignore_already_exists ? true : null

  # Destroy-time guard: PREVENT fails any destroy while set. Null falls
  # back to the provider default (DELETE).
  deletion_policy = var.spec.deletion_policy != null && var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}

# Optional user-managed key. Created only when spec.user_managed_key is
# present — keyless patterns (Workload Identity, impersonation, federation)
# are the recommended default. In the generate flow the private key is
# returned once at creation and marked sensitive in state; in the upload
# flow (public_key_data set) GCP never sees a private key at all.
resource "google_service_account_key" "main" {
  count = local.create_key ? 1 : 0

  service_account_id = google_service_account.main.name

  # Generate-flow shape. Nulls fall back to GCP defaults (2048-bit RSA,
  # JSON credentials file, X.509 PEM public key).
  key_algorithm    = var.spec.user_managed_key.algorithm != null && var.spec.user_managed_key.algorithm != "" ? var.spec.user_managed_key.algorithm : null
  private_key_type = var.spec.user_managed_key.private_key_type != null && var.spec.user_managed_key.private_key_type != "" ? var.spec.user_managed_key.private_key_type : null
  public_key_type  = var.spec.user_managed_key.public_key_type != null && var.spec.user_managed_key.public_key_type != "" ? var.spec.user_managed_key.public_key_type : null

  # Upload-flow shape: the caller's own public key (base64 X.509 PEM);
  # mutually exclusive with the *_type args above (spec CEL enforces it).
  public_key_data = var.spec.user_managed_key.public_key_data != null && var.spec.user_managed_key.public_key_data != "" ? var.spec.user_managed_key.public_key_data : null

  # Rotation trigger: any change to this map replaces the key.
  keepers = var.spec.user_managed_key.keepers

  deletion_policy = var.spec.user_managed_key.deletion_policy != null && var.spec.user_managed_key.deletion_policy != "" ? var.spec.user_managed_key.deletion_policy : null
}

# Project-level role grants. google_project_iam_member is ADDITIVE: each grant
# manages exactly one (role, member) pair and never clobbers other members'
# bindings on the same role — safe to compose with grants made elsewhere.
resource "google_project_iam_member" "project_roles" {
  for_each = toset(local.project_iam_roles)

  # The grant must target the account's home project explicitly; deriving it
  # from the created account (rather than re-reading spec) keeps the grant
  # correct when the project fell back to the provider default.
  project = google_service_account.main.project
  role    = each.value
  member  = "serviceAccount:${local.service_account_email}"

  depends_on = [google_service_account.main]
}

# Organization-level role grants (additive, same member semantics as above).
# Org-scope roles affect every project under the organization.
resource "google_organization_iam_member" "org_roles" {
  for_each = toset(local.org_iam_roles)

  org_id = var.spec.org_id
  role   = each.value
  member = "serviceAccount:${local.service_account_email}"

  depends_on = [google_service_account.main]
}
