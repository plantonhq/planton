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
  description = "GcpKmsAutokeyConfig specification"
  type = object({
    # The folder or project Autokey is configured on. Omit for the
    # provider's default project.
    scope = optional(object({
      # Project configuration: a literal project ID or a GcpProject
      # reference. Overrides any folder configuration above the project.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      project_id = optional(string, "")

      # Folder configuration: the folder's numeric ID -- a literal or a
      # GcpFolder reference. Every project beneath it inherits the
      # configuration unless the project sets its own.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      folder_id = optional(string, "")
    }))

    # How Autokey picks the project a new key is created in:
    #   "RESOURCE_PROJECT"      -- same-project storage: the key lives beside
    #                              the resource it protects
    #   "DEDICATED_KEY_PROJECT" -- dedicated-project storage: every key for
    #                              the folder lives in key_project (folder
    #                              configurations only)
    #   "DISABLED"              -- Autokey is off for the scope, overriding a
    #                              parent folder that has it on
    # Empty sends nothing; on a folder with key_project set Google then
    # treats the folder as dedicated-project storage.
    key_project_resolution_mode = optional(string, "")

    # The dedicated key project for a folder configuration: a literal
    # project ID or a GcpProject reference. The module sends Google's
    # projects/{id} form and enables the Cloud KMS API there. Its Cloud KMS
    # service agent needs roles/cloudkms.admin on it before the first key
    # handle is requested.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    key_project = optional(string, "")

    # What destroying this block does:
    #   "" / "DELETE" -- clears the configuration: Autokey is off for the
    #                    scope (existing keys stay)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the block leaves management and the configuration
    #                    stays in force
    deletion_policy = optional(string, "")
  })
}
