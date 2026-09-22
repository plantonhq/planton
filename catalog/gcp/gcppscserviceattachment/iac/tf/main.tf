# The producer half of Private Service Connect: one service attachment in
# front of the producer's internal load balancer (target_service, a regional
# forwarding rule), translating consumer traffic into the PSC NAT subnets.
# The connection policy (automatic or manual acceptance, accept and reject
# lists, reconciliation) changes in place; the name, region, and domain
# names recreate the attachment.
resource "google_compute_service_attachment" "this" {
  name        = local.attachment_name
  project     = local.project_id
  region      = var.spec.region
  description = local.description

  # The published load balancer and the NAT subnets consumer traffic lands
  # in. The provider stores nat_subnets as a set, so order never diffs.
  target_service = var.spec.target_service
  nat_subnets    = var.spec.nat_subnets

  # Who may connect. ACCEPT_MANUAL admits only the accept list (spec CEL
  # keeps the list off the automatic form); the reject list applies to both.
  connection_preference = var.spec.connection_preference

  dynamic "consumer_accept_lists" {
    for_each = local.consumer_accept_lists
    content {
      connection_limit  = consumer_accept_lists.value.connection_limit
      project_id_or_num = consumer_accept_lists.value.project_id_or_num
      network_url       = consumer_accept_lists.value.network_url
      endpoint_url      = consumer_accept_lists.value.endpoint_url
    }
  }

  consumer_reject_lists = length(var.spec.consumer_reject_lists) > 0 ? var.spec.consumer_reject_lists : null

  # Optional+Computed on the provider: sent only when the spec sets it, so
  # Google's default (false) is never fought.
  reconcile_connections = var.spec.reconcile_connections

  # Required by the API: the manifest states it either way (proto default
  # false is a real answer).
  enable_proxy_protocol = var.spec.enable_proxy_protocol

  domain_names = length(var.spec.domain_names) > 0 ? var.spec.domain_names : null

  # The tri-state propagated limit and its send-if-zero twin (locals.tf).
  propagated_connection_limit              = local.propagated_connection_limit
  send_propagated_connection_limit_if_zero = local.send_propagated_connection_limit_if_zero

  # Google's API currently ignores the flag; the module still sends the
  # stated intent so it takes effect the day the API honors it.
  show_nat_ips = var.spec.show_nat_ips ? true : null

  # What destroy does to a published service: DELETE (default), PREVENT
  # (refuse), or ABANDON (drop from state, keep serving).
  deletion_policy = local.deletion_policy
}
