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
  description = "GcpColabRuntime specification"
  type = object({
    # The GCP project the runtime lives in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Colab Enterprise region, e.g. "us-central1" -- the template's
    # region. Immutable.
    location = string

    # The runtime's id -- the last segment of its resource name. Lowercase
    # letters, digits, and hyphens. Defaults to metadata.name. Immutable.
    runtime_id = optional(string, "")

    # The template the runtime is built from: a GcpColabRuntimeTemplate
    # reference or a literal
    # projects/{project}/locations/{location}/notebookRuntimeTemplates/{id}.
    # Required: Colab Enterprise assigns every runtime from a template.
    # Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    runtime_template = string

    # The email of the user the runtime belongs to -- only this user can
    # connect notebooks to it. Immutable.
    runtime_user = string

    # The name shown in Colab Enterprise -- up to 128 characters. Defaults to
    # metadata.name. Immutable.
    display_name = optional(string, "")

    # What the runtime is for. Immutable.
    description = optional(string, "")

    # Whether the runtime should be running:
    #   "RUNNING" -- started (the default); compute bills
    #   "STOPPED" -- stopped; the disk stays and bills, compute does not
    # The block starts or stops the runtime to match on every apply. Unset
    # leaves the runtime running.
    desired_state = optional(string, "")

    # Upgrade the runtime to the latest Colab image whenever it is started
    # and Google reports it upgradable.
    auto_upgrade = optional(bool, false)

    # What happens to the runtime when this resource is destroyed:
    #   "" / "DELETE" -- the runtime and its disk are deleted
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the runtime leaves management and stays in GCP
    deletion_policy = optional(string, "")
  })
}
