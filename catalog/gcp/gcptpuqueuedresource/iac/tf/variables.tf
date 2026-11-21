variable "metadata" {
  description = "Catalog object metadata"
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
  description = "GcpTpuQueuedResource specification"
  type = object({
    # The GCP project the request and its nodes live in: a literal project
    # ID or a GcpProject reference. If omitted, the provider's default
    # project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The zone the capacity is requested in, e.g. "us-central1-a". Every
    # node is created here.
    zone = string

    # The request's id. Lowercase letters, digits, and hyphens, starting
    # with a letter. Defaults to metadata.name.
    queued_resource_id = optional(string, "")

    # The TPU nodes requested, provisioned together once capacity exists.
    node_specs = list(object({
      # The node's id once provisioned -- its name in the project, what
      # `gcloud compute tpus tpu-vm ssh` takes. Lowercase letters, digits, and
      # hyphens, starting with a letter. Required: Google's provider neither
      # reads back an id Google generates nor lets it change in place, so an
      # unnamed node would plan a replacement of the whole request on every
      # deploy.
      node_id = string

      # The node itself.
      node = object({
        # The TPU software image, matched to the accelerator generation, e.g.
        # "tpu-ubuntu2204-base" or "v2-alpha-tpuv5-lite".
        runtime_version = string

        # The slice: generation and chip or core count, e.g. "v2-8",
        # "v5litepod-8", "v6e-16". Unset: Google's default of "v2-8".
        accelerator_type = optional(string, "")

        # What the node is for.
        description = optional(string, "")

        # The node's network. Unset: the project's network and subnetwork named
        # "default", which many organizations never create -- a project without
        # them must set this.
        network_config = optional(object({
          # The VPC network: a GcpVpcNetwork reference or a literal
          # projects/{project}/global/networks/{name}. Unset: the project's network
          # named "default", which many organizations never create (or delete) --
          # a project without it must name a network here.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          network = optional(string, "")

          # The subnetwork in the zone's region: a GcpSubnetwork reference or a
          # literal projects/{project}/regions/{region}/subnetworks/{name}. Unset:
          # the subnetwork named "default" in the region, which exists only where
          # the project's "default" network does -- a project without it must
          # name a subnetwork here.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          subnetwork = optional(string, "")

          # Give the node's workers external IP addresses. Without them the
          # subnetwork needs Private Google Access (or Cloud NAT).
          enable_external_ips = optional(bool, false)

          # Let the workers forward packets with non-matching addresses.
          can_ip_forward = optional(bool, false)

          # The number of queues on the interface. Unset: Google's default.
          queue_count = optional(number, 0)
        }))
      })
    }))

    # What happens to the request when this resource is destroyed:
    #   "" / "DELETE" -- the request is deleted without force: Google refuses
    #                    while it is ACTIVE, PROVISIONING, or
    #                    WAITING_FOR_RESOURCES (see the message comment)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the request leaves management; its nodes keep
    #                    running (and billing)
    deletion_policy = optional(string, "")
  })
}
