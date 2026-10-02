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
  description = "GcpPlantonRunner specification"
  type = object({
    # The GCP project the runner is created in. Accepts a literal project ID
    # or a reference to a GcpProject resource. If omitted, the provider's
    # default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The GCP region the runner is deployed in, e.g. "us-central1". Deploy
    # the runner in the same region as the private endpoints it needs to
    # reach.
    region = string

    # The runner token that authorizes this runner to JOIN the control
    # plane. Create one with `planton runner token create` (or in the
    # console under Organization Settings -> Runner Tokens); on Planton, the
    # platform mints a token and writes it at exactly the managed-secret
    # reference this field names, before the infrastructure applies -- there
    # is no manual credential step. The token only gates joining and is
    # never the runner's identity: the runner receives its own individually
    # revocable identity when it registers itself on arrival, and revoking
    # this token never touches runners it already admitted. This is a
    # secret: supply it as a managed-secret reference, never inline
    # plaintext; it reaches the runner through Secret Manager, not through
    # any launch configuration.
    token = string

    # The control-plane endpoint the runner joins, as host:port. Leave
    # unset for Planton's hosted control plane (the runner's built-in
    # default); set it for a self-hosted instance (e.g.
    # "planton.example.com:443"). This is the one bootstrap coordinate the
    # join cannot deliver -- everything else (work queue, tunnel, API
    # endpoints) arrives in the join response.
    control_plane_endpoint = optional(string, "")

    # The runner build to deploy: an image tag of the official runner
    # container image. "latest" tracks the newest release; pin a specific
    # version tag for change control. New instances pull the tag on every
    # (re)start.
    runner_version = optional(string)

    # The container image repository the runner is pulled from. Override
    # only for air-gapped or mirrored registries hosting a copy of the
    # official image; the digest-identical mirror is your responsibility.
    image_repository = optional(string)

    # The service account the runner runs as -- its GCP identity when cloud
    # operations or IaC runs use keyless access instead of injected keys.
    # Accepts a literal email or a reference to a GcpServiceAccount
    # resource. When unset, the deployment creates a dedicated
    # permissionless service account so the identity seam always exists and
    # permissions can be granted later without replacing the runner --
    # deliberately never the project's Compute Engine default (which
    # typically carries broad project access the runner should not inherit).
    # Either way, the module grants exactly that account read access to the
    # runner's own token secret, nothing else.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    service_account = optional(string, "")

    # Private networking for OUTBOUND traffic: routes the runner's egress
    # into a VPC through Direct VPC egress, which is what lets it reach
    # private endpoints (a private GKE control plane, a private-IP
    # database). Only traffic to private ranges rides the VPC; the runner's
    # control-plane dial-out keeps its normal internet path. Omit when the
    # runner only needs to reach public endpoints.
    vpc_access = optional(object({
      # The VPC network. Accepts a literal name or a GcpVpcNetwork reference.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      network = string

      # The subnetwork the runner draws IPs from. Must be in the runner's
      # region.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      subnetwork = string

      # Network tags applied to the runner's traffic -- how VPC firewall
      # rules select its egress (e.g. a rule admitting the runner to a
      # private GKE control plane by tag).
      tags = optional(list(string), [])
    }))

    # CPU allocated to the runner instance, in vCPUs. The default 1
    # comfortably runs the runner's control loops plus typical IaC
    # operations; size up for large stacks or high operation concurrency.
    # Cloud Run admits 1, 2, 4, 6, or 8 vCPUs, and each pairs with a
    # minimum memory (see memory).
    cpu = optional(string)

    # Memory allocated to the runner instance, e.g. "512Mi" or "2Gi". The
    # default 512Mi pairs with the default cpu of 1. Cloud Run requires at
    # least 2Gi for 4 vCPUs and 4Gi for 6 or 8 vCPUs. Memory pressure shows
    # up as failed IaC operations mid-apply; when in doubt, size memory up
    # before cpu.
    memory = optional(string)
  })
}
