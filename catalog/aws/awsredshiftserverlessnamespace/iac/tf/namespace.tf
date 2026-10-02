# The namespace is the DATA plane of the serverless warehouse: the
# database, its admin credentials, the data-encryption key, the engine
# IAM roles, and audit log exports. Compute lives on
# AwsRedshiftServerlessWorkgroup nodes that attach to this namespace by
# name -- this module never creates or mutates a resource that deserves
# to be its own node (KMS keys and IAM roles attach by reference).
#
# Create-only in AWS: the namespace name and the first database's name.
# Everything else edits in place, including credential changes and KMS
# re-encryption (in place but long-running).
resource "aws_redshiftserverless_namespace" "this" {
  namespace_name = local.namespace_name

  # Empty keeps the AWS default first database ("dev"). Create-time
  # only -- additional databases are created with SQL, not here.
  db_name = var.spec.db_name != "" ? var.spec.db_name : null

  # Empty keeps the AWS default admin ("admin"). A serverless namespace
  # does not hard-require admin credentials at all -- IAM identities can
  # use temporary credentials without one.
  admin_username = var.spec.admin_username != "" ? var.spec.admin_username : null

  # The password contract (CEL enforces exactly one strategy): the
  # AWS-managed Secrets Manager secret (recommended -- no secret in
  # manifest or state) or a directly supplied password.
  # manage_admin_password is forwarded ONLY when true: an explicit
  # false conflicts with admin_user_password in the provider's
  # ConflictsWith machinery.
  manage_admin_password            = var.spec.manage_admin_password ? true : null
  admin_user_password              = var.spec.admin_user_password != "" ? var.spec.admin_user_password : null
  admin_password_secret_kms_key_id = var.spec.admin_password_secret_kms_key_id != "" ? var.spec.admin_password_secret_kms_key_id : null

  # Data encryption at rest. Empty keeps the AWS-owned Redshift service
  # key; switching keys later is an in-place but long-running
  # re-encryption.
  kms_key_id = var.spec.kms_key_id != "" ? var.spec.kms_key_id : null

  # IAM roles the serverless engine assumes for COPY/UNLOAD/Spectrum.
  # The default role must also be in iam_roles (an AWS requirement the
  # error message makes obvious enough to leave to the API).
  iam_roles            = length(var.spec.iam_roles) > 0 ? var.spec.iam_roles : null
  default_iam_role_arn = var.spec.default_iam_role_arn != "" ? var.spec.default_iam_role_arn : null

  # Audit log delivery to CloudWatch Logs. Empty exports nothing.
  log_exports = length(var.spec.log_exports) > 0 ? var.spec.log_exports : null

  tags = local.aws_tags
}
