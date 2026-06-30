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
  description = "Specification for the GCP Service Account"
  type = object({
    # The short unique ID for the service account (6-30 chars).
    # Used to form the email <service_account_id>@<project>.iam.gserviceaccount.com.
    # Required: must comply with GCP naming rules (lowercase letters, digits, hyphens).
    service_account_id = string

    # The GCP project ID in which the service account will be created.
    # Can be a literal value or a reference to a GcpProject resource.
    # If omitted, the provider default project is used.
    project_id = object({
      value = string
    })

    # Organization ID for organization-level IAM bindings.
    # Required if org_iam_roles is specified.
    org_id = optional(string)

    # Whether to create a JSON key for this service account.
    # Defaults to false (keyless is recommended for security).
    create_key = optional(bool)

    # List of IAM roles to grant at the project level.
    # Example: ["roles/logging.logWriter", "roles/storage.objectViewer"]
    project_iam_roles = optional(list(string))

    # List of IAM roles to grant at the organization level.
    # Requires org_id to be set.
    # Example: ["roles/resourcemanager.organizationViewer"]
    org_iam_roles = optional(list(string))
  })

  validation {
    condition     = length(var.spec.service_account_id) >= 6 && length(var.spec.service_account_id) <= 30
    error_message = "service_account_id must be between 6 and 30 characters."
  }

  validation {
    condition     = length(var.spec.org_iam_roles) == 0 || (var.spec.org_id != null && var.spec.org_id != "")
    error_message = "org_id must be specified when org_iam_roles is not empty."
  }
}
