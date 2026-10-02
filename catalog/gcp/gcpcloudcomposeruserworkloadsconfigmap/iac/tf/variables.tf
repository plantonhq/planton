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
  description = "GcpCloudComposerUserWorkloadsConfigMap specification"
  type = object({
    # GCP project of the Composer environment.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Region of the Composer environment (e.g., "us-central1").
    # Immutable after creation.
    region = string

    # The Composer environment the ConfigMap is delivered into. Resolves
    # to the environment's name. Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    environment = string

    # Name of the Kubernetes ConfigMap. Must be lowercase letters,
    # numbers, and hyphens; start with a letter; end with a letter or
    # number. Immutable after creation.
    config_map_name = string

    # The ConfigMap's key-value entries (plain configuration data).
    data = optional(map(string), {})

    # Deletion policy for the ConfigMap — what happens when this resource
    # is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the Kubernetes ConfigMap is removed from the
    #                environment; DAGs consuming it start failing
    #   "PREVENT" -- destroy FAILS; protects configuration live pipelines
    #                depend on from riding along with a stack teardown
    #   "ABANDON" -- the ConfigMap is removed from management but stays
    #                in the environment's cluster
    deletion_policy = optional(string, "")
  })
}
