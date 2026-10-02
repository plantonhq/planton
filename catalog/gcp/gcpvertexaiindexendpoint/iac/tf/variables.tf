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
  description = "GcpVertexAiIndexEndpoint specification"
  type = object({
    # GCP project where the index endpoint will be created.
    # If omitted, the endpoint is created in the provider's default
    # project (from the credential or ambient configuration).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Region where the index endpoint will be created (e.g.,
    # "us-central1"). Indexes deployed onto it must live in the same
    # region. Immutable after creation.
    location = string

    # Display name of the index endpoint (up to 128 UTF-8 characters).
    # The primary human-readable identifier; the numeric resource ID is
    # GCP-assigned.
    display_name = string

    # Description of the index endpoint.
    description = optional(string, "")

    # If true, deployed indexes are queryable through a public domain
    # name (the public_endpoint_domain_name output). Mutually exclusive
    # with network and private_service_connect_config. Immutable.
    public_endpoint_enabled = optional(bool, false)

    # VPC network to peer the endpoint into (private queries via Private
    # Services Access). The Vertex AI API expects the RELATIVE network
    # form projects/{project}/global/networks/{name} (with {project}
    # preferably the project NUMBER); both IaC modules normalize a
    # compute self-link URL (the GcpVpcNetwork reference's canonical
    # output) to that relative form. Requires Private Services Access
    # configured on the network. Mutually exclusive with
    # public_endpoint_enabled and private_service_connect_config.
    # Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    network = optional(string, "")

    # Private Service Connect configuration. When present, consumers
    # reach deployed indexes through a PSC service attachment. Mutually
    # exclusive with public_endpoint_enabled and network. Immutable.
    private_service_connect_config = optional(object({
      # Must be true when this block is present — the API's enablement flag
      # for PSC on the endpoint. Modeled explicitly (not inferred from block
      # presence) because it is the API's own contract field. Immutable.
      enable_private_service_connect = optional(bool, false)

      # Projects allowed to create forwarding rules targeting this
      # endpoint's service attachment. Each entry is a GCP project ID or
      # project number. Immutable.
      project_allowlist = optional(list(string), [])

      # PSC endpoints Vertex AI creates automatically in consumer
      # projects/networks (instead of consumers wiring forwarding rules by
      # hand). The provider models this field on index endpoints, but the
      # live API does NOT honor it there: a create carrying automation
      # configs succeeds while the stored endpoint omits them and no
      # consumer-side endpoint is ever provisioned (API-verified against a
      # live index endpoint; the provider documents the field as used by
      # online inference endpoints only). Because the PSC block is
      # immutable, that silent drop would surface to users as a perpetual
      # replacement diff on every re-plan — so this field is refused by
      # validation until Google extends automation to vector search. Use
      # project_allowlist with consumer-managed forwarding rules. Immutable.
      psc_automation_configs = optional(list(object({
        # VPC network where the PSC endpoint is created, as the full
        # relative resource name projects/{project}/global/networks/{name}
        # (the format the Vertex AI API requires) — a GcpVpcNetwork
        # reference's self-link output is normalized to that form by both
        # IaC modules. Immutable.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        network = string

        # Project in which the PSC endpoint (forwarding rule) is created —
        # a project ID; a GcpProject reference resolves to it. Immutable.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        project_id = string
      })), [])
    }))

    # User-defined labels to organize the index endpoint (cost
    # attribution, team ownership, environment tagging). Keys and values
    # must follow GCP label rules: lowercase letters, digits,
    # underscores, and dashes, at most 63 characters. Merged with the
    # platform's attribution labels; on key conflicts the platform
    # labels win. Mutable in place.
    labels = optional(map(string), {})

    # Cloud KMS key for customer-managed encryption at rest (CMEK) of
    # data on the endpoint's serving replicas, as the full key resource
    # path
    # projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
    # — a GcpKmsKey reference resolves to it. The key must live in the
    # same region as the endpoint, and the Vertex AI service agent needs
    # roles/cloudkms.cryptoKeyEncrypterDecrypter on it. If omitted, data
    # is encrypted with Google-managed keys. Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # Deletion policy for the index endpoint — what happens when this
    # resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the endpoint is deleted; every index deployed onto
    #                it stops serving
    #   "PREVENT" -- destroy FAILS; a guard for the serving surface all
    #                of this endpoint's deployed indexes depend on
    #   "ABANDON" -- the endpoint is removed from management but left
    #                standing (and billing for its replicas) in GCP
    deletion_policy = optional(string, "")
  })
}
