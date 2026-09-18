# A database user on a Cloud SQL instance. Users are first-class nodes: one
# per application/service with its own credential, instead of sharing the
# instance's admin user.
#
# No API enablement here: the instance this user lives on cannot exist
# without sqladmin.googleapis.com already enabled (its own module enables
# it).
#
# BUILT_IN users authenticate with the password below (rotatable in place —
# updating the field updates the credential without recreating the user).
# CLOUD_IAM_* users authenticate through IAM and carry no password at all;
# on PostgreSQL the instance must first set the database flag
# "cloudsql.iam_authentication" = "on".
resource "google_sql_user" "this" {
  # Derived in locals: the IAM service-account form when the user is a
  # service account, the spec's user_name otherwise.
  name     = local.user_name
  project  = local.project_id
  instance = var.spec.instance

  # Sensitive: flows from the spec's secret-annotated field; never exported
  # in outputs.
  password = local.password

  type = local.user_type

  # MySQL-only user@host scoping; null on other engines.
  host = local.host

  # MySQL 8+ / PostgreSQL: roles granted at creation (custom roles must
  # already exist in the database).
  database_roles = length(var.spec.database_roles) > 0 ? var.spec.database_roles : null

  # DELETE (default) drops the user; ABANDON removes it from IaC
  # management — the documented workaround when owned objects block a
  # PostgreSQL drop; PREVENT fails destroying plans.
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  dynamic "password_policy" {
    for_each = local.password_policy != null ? [local.password_policy] : []
    content {
      allowed_failed_attempts      = password_policy.value.allowed_failed_attempts
      password_expiration_duration = password_policy.value.password_expiration_duration != "" ? password_policy.value.password_expiration_duration : null
      enable_failed_attempts_check = password_policy.value.enable_failed_attempts_check
      enable_password_verification = password_policy.value.enable_password_verification
    }
  }
}
