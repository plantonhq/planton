variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Specification for the GCP Cloud SQL user"
  type = object({
    # The GCP project that owns the Cloud SQL instance. The CLI's tfvars
    # converter resolves StringValueOrRef fields to their literal string
    # before the module runs, so this arrives as a plain string.
    # If empty, the provider's default project is used (see locals.tf).
    project_id = optional(string, "")

    # Name of the Cloud SQL instance this user lives on; arrives as a plain
    # string after ref resolution. Immutable.
    instance = string

    # Login name (BUILT_IN) or IAM principal (IAM types). Exactly one of
    # user_name or service_account. Immutable.
    user_name = optional(string, "")

    # Service account email (a plain string after ref resolution) for a
    # CLOUD_IAM_SERVICE_ACCOUNT user; the module derives the database
    # username from it. Exactly one of user_name or service_account.
    service_account = optional(string, "")

    # Password for BUILT_IN users. Mutable (rotates in place). Never set
    # for IAM types (spec CEL enforces pre-deploy).
    password = optional(string, "")

    # BUILT_IN (middleware default), CLOUD_IAM_USER,
    # CLOUD_IAM_SERVICE_ACCOUNT, or CLOUD_IAM_GROUP. Immutable.
    type = optional(string, "BUILT_IN")

    # MySQL only: the host the user may connect from (user@host semantics).
    # Immutable.
    host = optional(string, "")

    # Per-user password policy (BUILT_IN only).
    password_policy = optional(object({
      allowed_failed_attempts      = optional(number, null)
      password_expiration_duration = optional(string, "")
      enable_failed_attempts_check = optional(bool, false)
      enable_password_verification = optional(bool, false)
    }), null)

    # MySQL 8+ / PostgreSQL: database roles granted at creation.
    database_roles = optional(list(string), [])

    # Engine-side teardown behavior: DELETE, PREVENT, or ABANDON. ABANDON
    # is the documented answer for PostgreSQL users that still own
    # database objects.
    deletion_policy = optional(string, "")
  })

  validation {
    condition     = (var.spec.user_name != "") != (var.spec.service_account != "")
    error_message = "Set exactly one of user_name or service_account."
  }

  validation {
    condition     = var.spec.service_account == "" || var.spec.type == "CLOUD_IAM_SERVICE_ACCOUNT"
    error_message = "service_account requires type CLOUD_IAM_SERVICE_ACCOUNT."
  }
}
