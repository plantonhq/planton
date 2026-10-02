variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "GcpCloudSqlUser specification"
  type = object({
    # The GCP project that owns the Cloud SQL instance.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Cloud SQL instance this user lives on. Accepts the instance name or
    # a reference to a GcpCloudSql resource. Immutable — a user cannot move
    # between instances.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    instance = string

    # The user name. Immutable. For BUILT_IN users this is the login name
    # ("orders-app"). For IAM types it is the IAM principal: the user's
    # full email for CLOUD_IAM_USER, the group email for CLOUD_IAM_GROUP,
    # and for CLOUD_IAM_SERVICE_ACCOUNT the service account's email with
    # the ".gserviceaccount.com" suffix dropped
    # ("ci-runner@my-project.iam") — the form Cloud SQL stores on
    # PostgreSQL; MySQL keeps only the part before "@". Both engines apply
    # that normalization for CLOUD_IAM_SERVICE_ACCOUNT, so a pasted full
    # email also works. Prefer service_account for the service-account
    # case: it wires the identity by reference instead of by hand. Exactly
    # one of user_name or service_account is set.
    user_name = optional(string, "")

    # Login password for a BUILT_IN user. Mutable — updating it rotates the
    # credential in place. Subject to the instance's password validation
    # policy. Never set for IAM-authenticated types.
    password = optional(string, "")

    # Authentication type. BUILT_IN (default) uses username + password; the
    # CLOUD_IAM_* types authenticate through IAM without a stored password.
    # Immutable.
    type = optional(string)

    # MySQL only: the host the user may connect from (classic MySQL
    # user@host semantics, e.g. "%" for any host or "10.0.0.0/8"). Leave
    # empty for PostgreSQL and SQL Server. Immutable.
    host = optional(string, "")

    # Per-user password policy (BUILT_IN users only), layered on top of the
    # instance-level password validation policy.
    password_policy = optional(object({
      # Number of failed login attempts after which the user is locked
      # (requires enable_failed_attempts_check).
      allowed_failed_attempts = optional(number)

      # Password lifetime as a seconds duration string, e.g. "2592000s" (30
      # days). After expiry the user must change the password to log in.
      password_expiration_duration = optional(string, "")

      # Whether failed login attempts are counted toward
      # allowed_failed_attempts.
      enable_failed_attempts_check = optional(bool, false)

      # MySQL only: require the current password when changing the password.
      enable_password_verification = optional(bool, false)
    }))

    # MySQL 8+ / PostgreSQL only: database roles granted to the user at
    # creation — predefined Cloud SQL roles (e.g. "cloudsqlsuperuser") or
    # custom roles already created in the database.
    database_roles = optional(list(string), [])

    # Engine-side teardown behavior. "DELETE" (default) drops the user;
    # "PREVENT" fails any plan that would drop it; "ABANDON" removes it
    # from IaC management while leaving it on the instance. ABANDON is the
    # documented answer for PostgreSQL users that cannot be dropped while
    # they still own database objects.
    deletion_policy = optional(string, "")

    # The service account this IAM database user represents, wired by
    # reference (a GcpServiceAccount's email) or as a literal email. Both
    # engines derive the database username Cloud SQL expects — the email
    # with ".gserviceaccount.com" dropped (Cloud SQL stores exactly that on
    # PostgreSQL, and only the part before "@" on MySQL) — so a chart never
    # hand-builds the form. Requires type CLOUD_IAM_SERVICE_ACCOUNT and the
    # instance's "cloudsql.iam_authentication" flag; the account still
    # needs roles/cloudsql.instanceUser (and roles/cloudsql.client to
    # connect) on the project, granted separately. Alternative to
    # user_name: exactly one of the two is set. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    service_account = optional(string, "")
  })
}
