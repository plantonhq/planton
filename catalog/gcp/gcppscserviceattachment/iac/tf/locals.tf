locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain; an empty string would be sent
  # verbatim and rejected by the API.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The cloud-side name defaults to metadata.name when the spec leaves
  # attachment_name empty -- the same naming basis every kind uses.
  attachment_name = var.spec.attachment_name != "" ? var.spec.attachment_name : var.metadata.name

  description = var.spec.description != "" ? var.spec.description : null

  # The accept list: each consumer names exactly one of a project, a
  # network, or an endpoint (spec CEL); the provider takes each arm as its
  # own optional argument with "" as the unset default, so the three are
  # sent as null when empty. The tfvars converter has already flattened the
  # project and network references to plain strings.
  consumer_accept_lists = [
    for c in var.spec.consumer_accept_lists : {
      connection_limit  = c.connection_limit
      project_id_or_num = c.project_id != "" ? c.project_id : null
      network_url       = c.network != "" ? c.network : null
      endpoint_url      = c.endpoint_url != "" ? c.endpoint_url : null
    }
  ]

  # propagated_connection_limit is tri-state in the spec: unset (null) lets
  # Google apply its default of 250; an explicit 0 must reach the API as 0,
  # which the provider only sends when its send_..._if_zero twin is true --
  # derived here from the spec value, never a field of its own (PARITY with
  # the Pulumi module).
  propagated_connection_limit              = var.spec.propagated_connection_limit
  send_propagated_connection_limit_if_zero = var.spec.propagated_connection_limit != null && var.spec.propagated_connection_limit == 0 ? true : null

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
