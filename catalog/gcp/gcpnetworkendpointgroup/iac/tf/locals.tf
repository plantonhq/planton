locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain; an empty string would be sent
  # verbatim and rejected by the API.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The scope selector. An empty zone builds the GLOBAL internet network
  # endpoint group; a zone name builds the ZONAL one. Exactly one of the two
  # group resources in main.tf exists (count guards), and outputs.tf picks
  # whichever was created.
  is_zonal = var.spec.zone != null && var.spec.zone != ""

  # The cloud-side name defaults to metadata.name when the spec leaves
  # neg_name empty -- the same naming basis every kind uses.
  neg_name = var.spec.neg_name != "" ? var.spec.neg_name : var.metadata.name

  description = var.spec.description != "" ? var.spec.description : null

  # The zonal resource defaults the type to GCE_VM_IP_PORT itself, so an
  # empty spec value is sent as null and Google's default stands; the global
  # resource requires the type (spec CEL guarantees it is set there).
  network_endpoint_type = var.spec.network_endpoint_type != "" ? var.spec.network_endpoint_type : null

  subnetwork = var.spec.subnetwork != "" ? var.spec.subnetwork : null

  # Endpoint fields are tri-state where the API distinguishes unset from
  # zero (port) and empty strings mean "not this arm" (instance, ip_address,
  # fqdn); the tfvars converter has already flattened the instance
  # reference to the instance name.
  zonal_endpoints = [
    for e in var.spec.endpoints : {
      instance   = e.instance != "" ? e.instance : null
      ip_address = e.ip_address != "" ? e.ip_address : null
      port       = e.port
    }
  ]

  global_endpoints = {
    for i, e in var.spec.endpoints : tostring(i) => {
      ip_address = e.ip_address != "" ? e.ip_address : null
      fqdn       = e.fqdn != "" ? e.fqdn : null
      # Global endpoints require a port; the group's default_port is the
      # fallback the spec allows.
      port = e.port != null ? e.port : var.spec.default_port
    }
  }

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
