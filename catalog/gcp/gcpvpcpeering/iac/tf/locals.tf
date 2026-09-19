locals {
  # The peering's name defaults to metadata.name when the spec leaves it
  # empty -- the same naming basis every kind uses. In the routes-config form
  # it names the EXISTING peering.
  peering_name = var.spec.peering_name != "" ? var.spec.peering_name : var.metadata.name

  # The form is selected by presence: a peer network means this side CREATES
  # the peering (google_compute_network_peering); none means it manages the
  # route exchange of a peering that already exists under peering_name
  # (google_compute_network_peering_routes_config).
  is_create_form = var.spec.peer_network != ""

  # The two public-IP subnet-route flags carry proto defaults (true / false)
  # that the manifest loader applies before either engine runs; the coalesce
  # is the same rule for a tfvars file written by hand. Both forms send them
  # always, so the manifest states the route exchange in full (PARITY with
  # the Pulumi module).
  export_subnet_routes_with_public_ip = coalesce(var.spec.export_subnet_routes_with_public_ip, true)
  import_subnet_routes_with_public_ip = coalesce(var.spec.import_subnet_routes_with_public_ip, false)

  # The routes-config resource addresses the network by NAME within a
  # project (unlike the peering, which takes the self link), so the name and
  # the project are taken from the self link when it carries them; a bare
  # name falls through unchanged and the project defers to the provider.
  network_segments = split("/", var.spec.network)
  network_name     = element(local.network_segments, length(local.network_segments) - 1)
  network_project = (
    contains(local.network_segments, "projects")
    ? element(local.network_segments, index(local.network_segments, "projects") + 1)
    : null
  )
}
