locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain; an empty string would be sent
  # verbatim and rejected by the API.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # Display name falls back to the resource's metadata name so every account
  # is human-identifiable in the console even when the field is omitted.
  display_name = coalesce(var.spec.display_name, var.metadata.name)

  # Computed service account email (used for IAM bindings)
  service_account_email = google_service_account.main.email

  # The user_managed_key message's PRESENCE is the decision to create a key;
  # its fields default to GCP's own defaults ({} = 2048-bit RSA JSON key).
  # The block's own switch (on by default once declared) decides whether a
  # key exists; a null switch reads as on.
  create_key = var.spec.user_managed_key != null && coalesce(var.spec.user_managed_key.enabled, true)

  # Whether the account starts disabled (defaults to false = enabled)
  disabled = coalesce(var.spec.disabled, false)

  # Project IAM roles (filter empty strings)
  project_iam_roles = var.spec.project_iam_roles != null ? [
    for role in var.spec.project_iam_roles : role
    if role != ""
  ] : []

  # Organization IAM roles (filter empty strings)
  org_iam_roles = var.spec.org_iam_roles != null ? [
    for role in var.spec.org_iam_roles : role
    if role != ""
  ] : []
}
