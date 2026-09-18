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
  description = "GcpCloudComposerUserWorkloadsSecret specification"
  type = object({
    # GCP project of the Composer environment.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Region of the Composer environment (e.g., "us-central1").
    # Immutable after creation.
    region = string

    # The Composer environment the Secret is delivered into. Resolves to
    # the environment's name. Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    environment = string

    # Name of the Kubernetes Secret. Must be lowercase letters, numbers,
    # and hyphens; start with a letter; end with a letter or number.
    # Immutable after creation.
    secret_name = string

    # The Secret's key-value entries. Values MUST be base64-encoded
    # (Kubernetes Secret semantics — e.g. `echo -n 'postgresql://...' |
    # base64`); the API rejects raw values. The decoded material (Airflow
    # connection URIs, passwords, tokens) is never placed in stack
    # outputs, and the entries are held as secrets in IaC state.
    data = optional(map(string), {})

    # Deletion policy for the Secret — what happens when this resource is
    # destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the Kubernetes Secret is removed from the
    #                environment; DAGs consuming it start failing
    #   "PREVENT" -- destroy FAILS; protects credentials live pipelines
    #                depend on from riding along with a stack teardown
    #   "ABANDON" -- the Secret is removed from management but stays in
    #                the environment's cluster
    deletion_policy = optional(string, "")
  })
}
