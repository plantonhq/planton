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
  description = "GcpVertexAiEndpoint specification"
  type = object({
    # GCP project where the endpoint will be created.
    # If omitted, the endpoint is created in the provider's default
    # project (from the credential or ambient configuration).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Region where the endpoint will be created (e.g., "us-central1").
    # Immutable after creation.
    location = string

    # Display name of the endpoint (up to 128 UTF-8 characters).
    # This is the primary human-readable identifier for the endpoint.
    display_name = string

    # Description of the endpoint.
    description = optional(string, "")

    # VPC network for private endpoints via VPC peering.
    # Format: projects/{project}/global/networks/{network}
    # Requires Private Services Access configured on the VPC.
    # Mutually exclusive with private_service_connect_config.
    # Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    network = optional(string, "")

    # KMS key for customer-managed encryption (CMEK).
    # Format: projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
    # If not specified, Google-managed encryption is used.
    # Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # If true, the endpoint is exposed through a dedicated DNS name
    # (https://{endpointId}.{region}-{projectNumber}.prediction.vertexai.goog)
    # rather than the shared regional DNS. Dedicated endpoints provide
    # better performance, reliability, and traffic isolation.
    # Mutually exclusive with private_service_connect_config.
    dedicated_endpoint_enabled = optional(bool, false)

    # Private Service Connect configuration. When present, the endpoint
    # is exposed via a PSC service attachment rather than VPC peering.
    # Mutually exclusive with network and dedicated_endpoint_enabled.
    private_service_connect_config = optional(object({
      # Projects allowed to create forwarding rules targeting this endpoint's
      # service attachment. Each entry is a GCP project ID or project number.
      # If empty, any project in the same organization can connect.
      # Mutable in place.
      project_allowlist = optional(list(string), [])

      # PSC endpoints Vertex AI creates automatically in consumer
      # projects/networks (instead of consumers wiring forwarding rules by
      # hand). Online-prediction endpoints are exactly the surface Google
      # documents this automation for. Mutable in place.
      psc_automation_configs = optional(list(object({
        # VPC network where the PSC endpoint is created, as the full relative
        # resource name projects/{project}/global/networks/{name} (the format
        # the Vertex AI API requires) — a GcpVpcNetwork reference's self-link
        # output is normalized to that form by both IaC modules.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        network = string

        # Project in which the PSC endpoint (forwarding rule) is created — a
        # project ID; a GcpProject reference resolves to it.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        project_id = string
      })), [])

      # Whether Private Service Connect is on. Unset means on: declaring the
      # block has always meant enabling PSC, and this switch lets a manifest say
      # the opposite out loud while keeping the allowlist and automation
      # settings in place (the modules then render no PSC block, since the API
      # rejects the block with its flag off).
      enabled = optional(bool)
    }))

    # GCP endpoint name (numeric identifier, max 10 digits).
    # GCP requires Vertex AI endpoint names to be numeric only --
    # this is the resource ID in the fully qualified path
    # projects/{project}/locations/{location}/endpoints/{name}.
    #
    # If not specified, the IaC module derives a stable numeric identifier
    # from the resource's own identity (organization, environment, and
    # resource name), so the same manifest always produces the same
    # endpoint ID on either provisioning engine. Most users should omit
    # this field and use display_name for human identification.
    # Immutable after creation.
    #
    # Reservation caveat: GCP reserves a deleted endpoint's numeric ID —
    # destroying an endpoint and recreating the same resource identity
    # (or reusing the same explicit ID) fails with 409 ALREADY_EXISTS
    # until GCP releases the reservation. To recreate immediately, set a
    # different explicit endpoint_name or change the resource name.
    endpoint_name = optional(string, "")

    # User-defined labels to organize the endpoint (cost attribution,
    # team ownership, environment tagging). Keys and values must follow
    # GCP label rules: lowercase letters, digits, underscores, and dashes,
    # at most 63 characters. Merged with the platform's attribution labels;
    # on key conflicts the platform labels win. Mutable in place.
    labels = optional(map(string), {})

    # Request/response logging for online predictions: samples prediction
    # traffic into a BigQuery table for drift monitoring, debugging, and
    # audit. Mutable in place.
    request_response_logging_config = optional(object({
      # Enable request/response logging.
      enabled = optional(bool, false)

      # Fraction of requests to log, in the range (0, 1]. Sample down
      # (e.g. 0.05) for high-QPS endpoints to bound BigQuery cost; use 1.0
      # to capture everything on low-traffic endpoints.
      sampling_rate = optional(number, 0)

      # BigQuery destination for the logged requests/responses, up to 2000
      # characters. Accepted forms:
      #   - "bq://projectId" -- GCP creates a dataset named
      #     logging_<endpoint-display-name>_<endpoint-id> and a
      #     request_response_logging table inside it.
      #   - "bq://projectId.bqDatasetId" -- GCP creates the table in the
      #     given dataset (the dataset must exist).
      #   - "bq://projectId.bqDatasetId.bqTableId" -- fully specified; the
      #     dataset must exist and the table must not.
      #
      # A plain string (not a reference) because the bq:// URI scheme has no
      # matching stack output on the BigQuery kinds; compose by writing the
      # dataset's project and ID into the URI.
      bigquery_destination_uri = optional(string, "")
    }))

    # Traffic routing across the models deployed on this endpoint: a map
    # from a DeployedModel's ID (assigned when a model is deployed — an
    # operational step outside this resource) to the percentage of traffic
    # it receives. GCP requires the values to add up to exactly 100, and
    # rejects IDs that are not currently deployed — so leave this EMPTY on
    # an endpoint with no deployed models (an empty map means the endpoint
    # accepts no traffic). Mutable in place: update it to shift traffic
    # between model versions (canary/blue-green serving).
    traffic_split = optional(map(number), {})

    # What happens to the endpoint in GCP when this resource is destroyed.
    #   "DELETE"  -- (GCP's default when unset) the endpoint is deleted;
    #                GCP rejects the delete while models are still deployed
    #                to it, so undeploy first
    #   "PREVENT" -- destroy FAILS; protects a serving URL applications
    #                still call
    #   "ABANDON" -- the endpoint is removed from management but keeps
    #                serving in GCP (deployed models keep billing)
    deletion_policy = optional(string, "")
  })
}
