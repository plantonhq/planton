locals {
  # Honor the spec contract: an empty project_id falls back to the
  # provider's default project. Passing null (instead of "") lets the
  # google provider resolve its own project from configuration.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The membership ID defaults to metadata.name and the location to
  # "global" -- identical to the Pulumi module.
  membership_id   = var.spec.membership_id != "" ? var.spec.membership_id : var.metadata.name
  location        = var.spec.location != "" ? var.spec.location : "global"
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # Google's resource link for a GKE cluster. The spec carries the
  # cluster's ID ("projects/{p}/locations/{l}/clusters/{name}"); the
  # module sends the "//container.googleapis.com/" form explicitly so both
  # engines send the same string Google stores.
  gke_cluster_resource_link = (
    var.spec.gke_cluster == "" ? null :
    startswith(var.spec.gke_cluster, "//") ? var.spec.gke_cluster :
    "//container.googleapis.com/${trimprefix(var.spec.gke_cluster, "/")}"
  )

  # The same planton-ai_* label set the Pulumi module applies. User labels
  # merge in first so the platform attribution labels can never be
  # clobbered by a spec label with the same key.
  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = var.metadata.name
    "planton-ai_kind"     = "gcpgkefleetmembership"
  }

  org_label = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "planton-ai_organization" = var.metadata.org } : {}

  env_label = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "planton-ai_environment" = var.metadata.env } : {}

  id_label = (
    var.metadata.id != null && var.metadata.id != ""
  ) ? { "planton-ai_id" = var.metadata.id } : {}

  final_labels = merge(var.spec.labels, local.base_labels, local.org_label, local.env_label, local.id_label)
}
