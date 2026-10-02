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
  description = "GcpProject specification"
  type = object({
    # Unique project ID for the GCP project. Must be 6-30 characters long,
    # contain only lowercase letters, digits, and hyphens, must start with
    # a letter, and cannot end with a hyphen. Globally unique across all of
    # GCP and IMMUTABLE — deleted project IDs stay reserved for up to 30
    # days, so they cannot be reused quickly.
    project_id = string

    # Human-readable display name shown in the console (4-30 characters).
    # Mutable, unlike project_id. If not specified, defaults to
    # metadata.name.
    display_name = optional(string, "")

    # The type of parent node the project is created under. Changing the
    # parent migrates the project within the hierarchy. May be left empty
    # when folder_id names the parent.
    parent_type = optional(string, "")

    # Organization ID or Folder ID (numeric string) matching parent_type.
    # For a folder declared in the same chart, prefer folder_id (a
    # reference) and leave this empty.
    parent_id = optional(string, "")

    # Billing account ID in the form "0123AB-4567CD-89EFGH".
    # Strongly recommended for any project that will use billable services.
    # The deploying identity needs roles/billing.user on the account.
    billing_account_id = optional(string, "")

    # User labels merged onto the project beneath the platform's attribution
    # labels (platform keys win on conflicts). Project labels are the
    # primary cost-allocation dimension in billing exports.
    # Keys/values: lowercase letters, digits, underscores, hyphens.
    labels = optional(map(string), {})

    # Resource Manager tags bound to the project at CREATE TIME only
    # (tagKeys/{id} -> tagValues/{id}, the `name` outputs of GcpTagKey and
    # GcpTagValue). Tags drive org policies and IAM conditions. Changing this
    # after creation recreates the project — for tags on an existing project,
    # bind tag values with GcpTagBinding instead, which attaches and detaches
    # without touching the project.
    tags = optional(map(string), {})

    # Whether GCP auto-creates the "default" VPC network in the new project.
    # Defaults to false: deleting the auto-created network is a standard
    # security-hardening step, and explicit GcpVpcNetwork resources are the
    # composable path. Note the project still needs one network slot of
    # quota available even when false (the network exists momentarily).
    auto_create_network = optional(bool)

    # List of Cloud APIs to enable at project creation
    # (e.g. "compute.googleapis.com"). Individual component kinds also
    # enable the APIs they need, so this is a convenience for pre-warming a
    # known set. Each entry must end with ".googleapis.com".
    enabled_apis = optional(list(string), [])

    # What destroying this resource does to the project:
    #   DELETE (default): the project is shut down (30-day pending-deletion
    #     window during which it can be restored).
    #   PREVENT: destroy fails — protection for shared foundation projects.
    #   ABANDON: the resource is removed from state and the project lives
    #     on unmanaged — the safe hand-off when ownership moves elsewhere.
    deletion_policy = optional(string, "")

    # The folder the project lives in, by reference: a GcpFolder resource
    # (its folder_id output) or the folder's numeric ID as a literal. This is
    # how a chart places a project inside a folder it also declares -- the
    # project waits for the folder to exist. When set it IS the parent:
    # leave parent_id empty and parent_type empty (or `folder`). Changing it
    # moves the project into the new folder in place; nothing is recreated,
    # but the IAM and organization policies inherited from the old folder
    # stop applying and the new folder's start.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    folder_id = optional(string, "")
  })
}
