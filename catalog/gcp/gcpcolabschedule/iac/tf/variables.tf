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
  description = "GcpColabSchedule specification"
  type = object({
    # The GCP project the schedule lives in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Vertex AI region the schedule and its runs live in, e.g.
    # "us-central1". Changing it needs a new schedule.
    location = string

    # The schedule's name in the console -- up to 128 characters. Defaults
    # to metadata.name.
    display_name = optional(string, "")

    # When runs launch, in cron syntax, optionally prefixed with a time zone:
    # "0 6 * * *" (06:00 UTC daily) or "TZ=America/New_York 0 9 * * 1-5"
    # (09:00 New York time on weekdays).
    cron = string

    # How many runs may be STARTED at the same time. A run that would exceed
    # it is skipped (or queued, with allow_queueing). This limits launching,
    # not how long a run executes.
    max_concurrent_run_count = optional(number, 0)

    # Queue a run that hits max_concurrent_run_count instead of skipping it.
    allow_queueing = optional(bool, false)

    # The earliest a run may launch, RFC 3339 (e.g. "2026-10-01T00:00:00Z").
    # Unset: the schedule's creation time. Sent only when set.
    start_time = optional(string, "")

    # After this time no new runs launch and the schedule completes, RFC
    # 3339. Unset: no end.
    end_time = optional(string, "")

    # Complete the schedule after this many runs have launched. Unset (0):
    # runs keep launching until the schedule is paused, ended, or deleted.
    max_run_count = optional(number, 0)

    # How many runs may be in a non-terminal state at once -- Google applies
    # it to pipeline schedules only. Unset (0): no limit beyond
    # max_concurrent_run_count.
    max_concurrent_active_run_count = optional(number, 0)

    # Whether the schedule launches runs:
    #   "ACTIVE" -- runs launch on the cron (the default)
    #   "PAUSED" -- nothing launches; resume by setting ACTIVE again
    # The block pauses or resumes the schedule to match on every apply.
    desired_state = optional(string, "")

    # The Colab Enterprise notebook run each tick launches. Exactly one of
    # notebook_execution_job or pipeline_job. A change replaces the
    # schedule.
    notebook_execution_job = optional(object({
      # The run's name in the console -- up to 128 characters. Defaults to the
      # schedule's display name.
      display_name = optional(string, "")

      # Run a notebook stored in Cloud Storage. Exactly one source.
      gcs_notebook_source = optional(object({
        # The notebook: gs://bucket/path/notebook.ipynb.
        uri = string

        # Pin an object generation (version) of the notebook. Unset: the current
        # version at each run.
        generation = optional(string, "")
      }))

      # Run a notebook from a Dataform repository (notebooks under version
      # control). Exactly one source.
      dataform_repository_source = optional(object({
        # The repository:
        # projects/{project}/locations/{location}/repositories/{repository}.
        dataform_repository_resource_name = string

        # Pin a commit. Unset: the repository's HEAD at each run.
        commit_sha = optional(string, "")
      }))

      # Run on a runtime template's machine, network, and image: a
      # GcpColabRuntimeTemplate reference or a literal
      # projects/{project}/locations/{location}/notebookRuntimeTemplates/{id}.
      # Exactly one of this or custom_environment_spec.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      notebook_runtime_template_resource_name = optional(string, "")

      # Run on a machine described here instead of a template. Exactly one of
      # this or notebook_runtime_template_resource_name.
      custom_environment_spec = optional(object({
        # The machine and accelerators.
        machine_spec = optional(object({
          # The Compute Engine machine type, e.g. "n1-standard-4". Immutable.
          machine_type = optional(string, "")

          # The accelerator, e.g. "NVIDIA_TESLA_T4", "NVIDIA_L4",
          # "NVIDIA_H100_80GB", "TPU_V5_LITEPOD".
          accelerator_type = optional(string, "")

          # How many accelerators the machine gets.
          accelerator_count = optional(number, 0)

          # Split each GPU into NVIDIA multi-instance partitions of this size
          # (e.g. "1g.10gb"); then accelerator_count should be 1. Immutable.
          gpu_partition_size = optional(string, "")

          # The TPU topology for a TPU accelerator, e.g. "2x2x1". Immutable.
          tpu_topology = optional(string, "")

          # Draw the machine from a Compute Engine reservation.
          reservation_affinity = optional(object({
            # "NO_RESERVATION", "ANY_RESERVATION", "SPECIFIC_RESERVATION",
            # "SPECIFIC_THEN_ANY_RESERVATION", or "SPECIFIC_THEN_NO_RESERVATION".
            reservation_affinity_type = string

            # The reservation label key; for a reservation by name,
            # "compute.googleapis.com/reservation-name".
            key = optional(string, "")

            # The reservation label values -- for a named reservation, its full
            # resource name.
            values = optional(list(string), [])

            # Draw from Google's shared Vertex AI capacity pool.
            use_reservation_pool = optional(bool, false)
          }))
        }))

        # The network the run joins.
        network_spec = optional(object({
          # Give the run a public internet path. Default false.
          enable_internet_access = optional(bool, false)

          # The VPC network: a GcpVpcNetwork reference or a literal
          # projects/{project}/global/networks/{name}.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          network = optional(string, "")

          # The subnetwork: a GcpSubnetwork reference or a literal
          # projects/{project}/regions/{region}/subnetworks/{name}.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          subnetwork = optional(string, "")
        }))

        # The run's persistent disk.
        persistent_disk_spec = optional(object({
          # "pd-standard" (Google's default), "pd-balanced", "pd-ssd", or
          # "pd-extreme".
          disk_type = optional(string, "")

          # Disk size in GB. Unset (0): Google's default of 100.
          disk_size_gb = optional(number, 0)
        }))
      }))

      # Run in a Vertex AI Workbench instance-based environment (Google's empty
      # workbench_runtime marker); still needs a template or a custom
      # environment for the machine. Sent only when true.
      workbench_runtime = optional(bool, false)

      # Where the executed notebook (with its outputs) is written:
      # a GcpGcsBucket reference (its gs:// URL) or a literal gs://bucket or
      # gs://bucket/prefix. The run's identity needs write access.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      gcs_output_uri = string

      # Run as this user (their email) -- their access applies. Exactly one of
      # execution_user or service_account.
      execution_user = optional(string, "")

      # Run as this service account: a GcpServiceAccount reference or a
      # literal email. The schedule's caller needs iam.serviceAccounts.actAs
      # on it. Exactly one of execution_user or service_account; the choice
      # for unattended production runs.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      service_account = optional(string, "")

      # The longest a run may execute, in seconds ending in "s" (e.g.
      # "3600s"). Unset: Google's default of 24 hours.
      execution_timeout = optional(string, "")

      # The Jupyter kernel to run the notebook with. Unset: the notebook's
      # default kernel.
      kernel_name = optional(string, "")

      # Labels on every run the schedule launches.
      labels = optional(map(string), {})

      # Customer-managed encryption key for the run: a GcpKmsKey reference or
      # a literal key path in the schedule's region.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key_name = optional(string, "")
    }))

    # The Vertex AI Pipelines run each tick launches. Exactly one of
    # notebook_execution_job or pipeline_job. Updates in place.
    pipeline_job = optional(object({
      # The run's name in the console -- up to 128 characters.
      display_name = optional(string, "")

      # The compiled pipeline definition as a JSON string -- the output of the
      # Kubeflow Pipelines SDK compiler. Set this or template_uri.
      pipeline_spec = optional(string, "")

      # Where the compiled pipeline is downloaded from when pipeline_spec is
      # empty -- an Artifact Registry pipeline template URI
      # (https://{region}-kfp.pkg.dev/{project}/{repository}/{template}/{tag}).
      template_uri = optional(string, "")

      # The run's parameters and output location.
      runtime_config = optional(object({
        # The Cloud Storage root the run's artifacts are written under
        # ({job_id}/{task_id}/{output_key}), e.g. gs://bucket/pipeline-root. The
        # pipeline's service account needs storage.objects.get and
        # storage.objects.create on it.
        gcs_output_directory = string

        # How the run reacts to a failed task:
        #   "PIPELINE_FAILURE_POLICY_FAIL_SLOW" -- running tasks finish, nothing
        #                                          new is scheduled
        #   "PIPELINE_FAILURE_POLICY_FAIL_FAST" -- the run stops at once
        failure_policy = optional(string, "")

        # Runtime parameters substituted into the pipeline's placeholders
        # (pipelines compiled with KFP SDK 1.9+ / the v2 DSL). Values are
        # strings.
        parameter_values = optional(map(string), {})
      }))

      # The service account the pipeline's steps run as: a GcpServiceAccount
      # reference or a literal email. Unset: the Compute Engine default
      # service account. The schedule's caller needs
      # iam.serviceAccounts.actAs on it.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      service_account = optional(string, "")

      # A VPC network the pipeline's workloads peer with (private services
      # access must already be configured on it): a GcpVpcNetwork reference or
      # a literal network path. Google wants
      # projects/{project_NUMBER}/global/networks/{name}; the modules resolve a
      # project ID to its number. Unset: no peering.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      network = optional(string, "")

      # Names of reserved IP ranges on the peered network the workloads may
      # use (e.g. "vertex-ai-ip-range"). Unset: any range on the network.
      reserved_ip_ranges = optional(list(string), [])

      # Validate every component before the run starts.
      preflight_validations = optional(bool, false)

      # Labels on every pipeline run. Google overrides the reserved
      # vertex-ai-pipelines-run-billing-id key.
      labels = optional(map(string), {})

      # Customer-managed encryption key for the run: a GcpKmsKey reference or
      # a literal key path in the schedule's region.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key_name = optional(string, "")

      # Private Service Connect interface: reach private services through a
      # network attachment instead of peering.
      psc_interface_config = optional(object({
        # The Compute Engine network attachment the run attaches to, in the
        # schedule's region and project (created beforehand).
        network_attachment = optional(string, "")

        # DNS peering so the run resolves private domains through another
        # project's Cloud DNS. The Vertex AI service agent needs roles/dns.peer
        # on each target project.
        dns_peering_configs = optional(list(object({
          # The DNS suffix to peer, ending with a dot, e.g.
          # "internal.example.com.".
          domain = string

          # The VPC network in target_project where the zone is visible: a
          # GcpVpcNetwork reference (its name) or a literal network name.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          target_network = string

          # The project hosting the Cloud DNS zone: a GcpProject reference or a
          # literal project ID.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          target_project = string
        })), [])
      }))
    }))

    # What happens to the schedule when this resource is destroyed:
    #   "" / "DELETE" -- the schedule is deleted (runs it already launched
    #                    and their outputs stay)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the schedule leaves management and keeps launching
    #                    runs
    deletion_policy = optional(string, "")
  })
}
