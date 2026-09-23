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
  description = "GcpColabRuntimeTemplate specification"
  type = object({
    # The GCP project the template lives in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Colab Enterprise region, e.g. "us-central1". Runtimes created from
    # the template run here. Changing it needs a new template.
    location = string

    # The template's id -- the last segment of its resource name. Lowercase
    # letters, digits, and hyphens. Defaults to metadata.name. Immutable.
    runtime_template_id = optional(string, "")

    # The name users see when they pick a runtime template -- up to 128
    # characters. Defaults to metadata.name. Mutable.
    display_name = optional(string, "")

    # What the template is for (e.g. "GPU runtime for the vision team").
    # Immutable.
    description = optional(string, "")

    # Labels on the template. The platform attribution labels are merged in
    # and win on key conflicts. A label change replaces the template.
    labels = optional(map(string), {})

    # The machine every runtime gets. Omit for Google's default machine.
    machine_spec = optional(object({
      # The Compute Engine machine type, e.g. "e2-standard-4" or
      # "n1-standard-8" (GPUs attach to N1 and the accelerator-optimized
      # families). Unset keeps Google's default.
      machine_type = optional(string, "")

      # The accelerator, e.g. "NVIDIA_TESLA_T4", "NVIDIA_L4",
      # "NVIDIA_TESLA_A100". Needs accelerator_count.
      accelerator_type = optional(string, "")

      # How many accelerators each runtime gets.
      accelerator_count = optional(number, 0)
    }))

    # The runtime's data disk (mounted as the home directory). Omit for
    # Google's default disk.
    data_persistent_disk_spec = optional(object({
      # "pd-standard", "pd-balanced", "pd-ssd", or "pd-extreme". Unset keeps
      # Google's default.
      disk_type = optional(string, "")

      # Disk size in GB, 10 to 65536. Needs disk_type.
      disk_size_gb = optional(number, 0)
    }))

    # The network the runtime joins. Omit for Google's default network with
    # internet access.
    network_spec = optional(object({
      # Give runtimes a public internet path. Turn off for runtimes that must
      # stay private; then the subnetwork needs Private Google Access.
      enable_internet_access = optional(bool, false)

      # The VPC network: a GcpVpcNetwork reference or a literal
      # projects/{project}/global/networks/{name}. Unset: the project's
      # default network.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      network = optional(string, "")

      # The subnetwork in the template's region: a GcpSubnetwork reference or
      # a literal projects/{project}/regions/{region}/subnetworks/{name}.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      subnetwork = optional(string, "")
    }))

    # Shut an idle runtime down after this long -- the main cost control for
    # interactive notebooks. A duration in seconds ending in "s": "0s"
    # disables idle shutdown; otherwise between 600s (10 minutes) and 86400s
    # (24 hours). Unset keeps Google's default idle shutdown.
    idle_timeout = optional(string, "")

    # Block the user's own Google credentials inside the runtime, so
    # notebooks act only as the runtime's service identity. Sent only when
    # set.
    euc_disabled = optional(bool, false)

    # Boot runtimes with Secure Boot (Shielded VM), refusing unsigned boot
    # components. Sent only when true.
    enable_secure_boot = optional(bool, false)

    # Compute Engine network tags applied to every runtime -- the handle VPC
    # firewall rules target. Immutable.
    network_tags = optional(list(string), [])

    # Customer-managed encryption key protecting the runtimes' disks: a
    # GcpKmsKey reference or a literal
    # projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
    # in the template's region. The Compute Engine and Vertex AI service
    # agents need roles/cloudkms.cryptoKeyEncrypterDecrypter on it. Mutable
    # (applies to runtimes created afterwards).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # The notebook software: environment variables, a post-startup script,
    # and the Colab image release. Mutable.
    software_config = optional(object({
      # Environment variables set in the notebook container.
      env = optional(list(object({
        # The variable name -- a valid C identifier.
        name = string

        # The value. $(VAR_NAME) expands a previously defined variable; $$(...)
        # escapes the expansion. Do not put secrets here -- the value is stored
        # in the template in plain text.
        value = optional(string, "")
      })), [])

      # A script run after the runtime starts (installing packages, mounting
      # data, configuring proxies).
      post_startup_script_config = optional(object({
        # The script itself, inline.
        post_startup_script = optional(string, "")

        # A URL the script is downloaded from, e.g.
        # gs://bucket/setup.sh or https://example.com/setup.sh.
        post_startup_script_url = optional(string, "")

        # When the script runs:
        #   "RUN_ONCE"                     -- on the first start only
        #   "RUN_EVERY_START"              -- on every start
        #   "DOWNLOAD_AND_RUN_EVERY_START" -- re-downloaded and run on every
        #                                     start (picks up script changes)
        post_startup_script_behavior = optional(string, "")
      }))

      # Which Colab image release the runtime boots.
      colab_image = optional(object({
        # The image release, e.g. "py310". Unset: the latest release.
        release_name = optional(string, "")
      }))
    }))

    # What happens to the template when this resource is destroyed:
    #   "" / "DELETE" -- the template is deleted (runtimes already created
    #                    from it keep running)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the template leaves management and stays in GCP
    deletion_policy = optional(string, "")
  })
}
