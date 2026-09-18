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
  description = "GcpWorkflow specification"
  type = object({
    # The GCP project to create the workflow in. Can be a literal project ID
    # or a reference to a GcpProject resource. If omitted, the provider's
    # default project is used. Immutable: changing it replaces the workflow.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The region the workflow runs in (e.g. us-central1). If omitted, the
    # provider's default region is used. Immutable: changing it replaces the
    # workflow.
    region = optional(string, "")

    # The workflow name in GCP. Defaults to metadata.name when left empty.
    # Immutable: changing it replaces the workflow (a new workflow with a
    # fresh execution history — running executions on the old name are
    # unaffected until it is deleted).
    workflow_name = optional(string, "")

    # What this workflow orchestrates — shown in the console list. At most
    # 1000 unicode characters (the provider's own documented cap).
    description = optional(string, "")

    # User labels attached to the workflow (merged with the platform's
    # standard labels by the module).
    labels = optional(map(string), {})

    # The workflow definition — the YAML (or JSON) source executed on each
    # run (https://cloud.google.com/workflows/docs/reference/syntax). The
    # API caps the size at 128KB. REQUIRED: a workflow without source has
    # never been deployable through the API — the provider still marks the
    # argument optional but its own 8.0.0 upgrade note says it "will become
    # REQUIRED ... to align with API constraints"; this spec models the API
    # truth today. Every source change deploys a NEW revision.
    source_contents = string

    # The service account the workflow's executions run AS — the identity
    # its HTTP calls and connector calls carry (format: bare email; the
    # module renders the projects/{project}/serviceAccounts/{email} form the
    # API stores). A literal email or a reference to a GcpServiceAccount
    # resource. If omitted, the project's default compute service account is
    # used — fine for experiments, wrong for production (grant a dedicated
    # account only the roles the steps need). Changing the service account
    # deploys a NEW revision.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    service_account = optional(string, "")

    # Customer-managed encryption key (CMEK) used to encrypt workflow
    # definitions and execution data at rest. The full crypto key resource
    # name (projects/{p}/locations/{l}/keyRings/{r}/cryptoKeys/{k}) — a
    # literal or a reference to a GcpKmsKey resource. Grant the Workflows
    # service agent roles/cloudkms.cryptoKeyEncrypterDecrypter on the key
    # BEFORE deploying, or the deploy fails. Omit for Google-managed
    # encryption.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    crypto_key = optional(string, "")

    # How much call detail execution history records:
    #   ""                          -- the API default (errors only)
    #   "CALL_LOG_LEVEL_UNSPECIFIED" -- explicit API default
    #   "LOG_ALL_CALLS"             -- every call and its result (debugging;
    #                                  highest log volume and cost)
    #   "LOG_ERRORS_ONLY"           -- failed calls only
    #   "LOG_NONE"                  -- nothing
    call_log_level = optional(string, "")

    # How much step-level detail execution history keeps:
    #   ""                                    -- the API default (basic)
    #   "EXECUTION_HISTORY_LEVEL_UNSPECIFIED" -- explicit API default
    #   "EXECUTION_HISTORY_BASIC"             -- step names and status
    #   "EXECUTION_HISTORY_DETAILED"          -- step inputs/outputs too
    #                                            (needed for step-level
    #                                            debugging in the console)
    execution_history_level = optional(string, "")

    # Environment variables visible to the workflow source via sys.get_env().
    # At most 20 entries (the API's own cap, enforced here); each value up to
    # 4KiB. Keys must be non-empty and must NOT start with "GOOGLE" or
    # "WORKFLOWS" (reserved prefixes the API rejects — key-shape rules live
    # here in the comment because map KEYS are not CEL-addressable). Changing
    # env vars deploys a NEW revision.
    user_env_vars = optional(map(string), {})

    # Resource manager tags bound at workflow creation, keyed
    # tagKeys/{tag_key_id} with values tagValues/{tag_value_id} — the
    # org-policy / cost-attribution tag system (distinct from labels).
    # Immutable: changing tags REPLACES the workflow (provider ForceNew) —
    # plan tag changes deliberately.
    resource_manager_tags = optional(map(string), {})

    # Guard rail: while true (the DEFAULT), destroying this resource FAILS
    # before touching the workflow. Set false explicitly to allow destroy.
    # Both IaC engines send the value explicitly on every apply so a
    # true -> false transition always reaches the engine (the
    # send-true-or-omit class would silently keep the guard up).
    deletion_protection = optional(bool)

    # Deletion policy — what happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the workflow and its execution history are deleted;
    #                running executions are cancelled
    #   "PREVENT" -- destroy FAILS (defense in depth alongside
    #                deletion_protection)
    #   "ABANDON" -- the workflow is removed from management but keeps
    #                running in GCP
    deletion_policy = optional(string, "")
  })
}
