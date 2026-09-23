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
  description = "GcpVertexAiTensorboard specification"
  type = object({
    # The GCP project the TensorBoard lives in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Vertex AI location (region) the TensorBoard lives in, e.g.
    # "us-central1". Training jobs that stream into it must run in the same
    # location. Immutable.
    location = string

    # Human-readable name shown in the console. Defaults to metadata.name.
    # Mutable in place.
    display_name = optional(string, "")

    # Free-text description of the TensorBoard.
    description = optional(string, "")

    # Labels on the TensorBoard. The platform attribution labels are merged
    # in and win on key conflicts.
    labels = optional(map(string), {})

    # Customer-managed encryption key protecting the TensorBoard and every
    # series logged to it: a GcpKmsKey reference or a literal
    # projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
    # in the same region. The Vertex AI Service Agent needs
    # roles/cloudkms.cryptoKeyEncrypterDecrypter on the key. Omit to use
    # Google-managed encryption. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # Experiments declared in the TensorBoard, each keyed by experiment_id.
    # Leave empty to let training jobs and the Vertex AI SDK create the
    # experiments they log to.
    experiments = optional(list(object({
      # The experiment's id -- the last segment of its resource name and the
      # name a training job or the Vertex AI SDK
      # (aiplatform.init(experiment=...)) logs under. 1-128 characters:
      # lowercase letters, digits, and hyphens. Unique within the TensorBoard.
      # Immutable.
      experiment_id = string

      # Human-readable name shown in the TensorBoard UI.
      display_name = optional(string, "")

      # Free-text description of the experiment.
      description = optional(string, "")

      # Labels on the experiment. The platform attribution labels are merged
      # in and win on key conflicts.
      labels = optional(map(string), {})

      # Where the experiment's data comes from, recorded as metadata -- e.g.
      # "custom training job" or a pipeline name. Informational; Google does
      # not act on it. Immutable.
      source = optional(string, "")

      # Runs declared up front, each keyed by run_id. Leave empty to let
      # training jobs create the runs they log to.
      runs = optional(list(object({
        # The run's id -- the last segment of its resource name. 1-128
        # characters: lowercase letters, digits, and hyphens. Unique within its
        # experiment. Immutable.
        run_id = string

        # Human-readable name -- Google requires one and requires it unique
        # among the experiment's runs. Defaults to run_id. Mutable in place.
        display_name = optional(string, "")

        # Free-text description of the run.
        description = optional(string, "")

        # Labels on the run. The platform attribution labels are merged in and
        # win on key conflicts.
        labels = optional(map(string), {})
      })), [])
    })), [])

    # What happens to the TensorBoard, its experiments, and its runs when
    # this resource is destroyed:
    #   "" / "DELETE" -- everything is deleted, logged series included
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- everything leaves management and stays in GCP
    deletion_policy = optional(string, "")
  })
}
