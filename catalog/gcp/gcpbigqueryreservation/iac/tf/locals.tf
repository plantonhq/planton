locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The reservation's name defaults to metadata.name -- the same fallback the
  # Pulumi module applies.
  reservation_name = var.spec.reservation_name != "" ? var.spec.reservation_name : var.metadata.name

  # Empty optionals become null so the provider omits them and Google's
  # defaults apply (location US, edition chosen by Google, concurrency
  # sized automatically, idle slots shared).
  location           = var.spec.location != "" ? var.spec.location : null
  edition            = var.spec.edition != "" ? var.spec.edition : null
  concurrency        = var.spec.concurrency > 0 ? var.spec.concurrency : null
  ignore_idle_slots  = var.spec.ignore_idle_slots ? true : null
  reservation_group  = var.spec.reservation_group != "" ? var.spec.reservation_group : null
  secondary_location = var.spec.secondary_location != "" ? var.spec.secondary_location : null
  deletion_policy    = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # Each assignment keyed by assignee, job type, and principal -- the
  # uniqueness the spec enforces -- with the assignee composed into the
  # projects/, folders/, or organizations/ form Google takes. Keys keep
  # declared order through assignment_keys, identical to the Pulumi module.
  assignment_list = [
    for a in var.spec.assignments : {
      assignee = (
        a.assignee.project_id != "" ? "projects/${a.assignee.project_id}" :
        a.assignee.folder_id != "" ? "folders/${a.assignee.folder_id}" :
        "organizations/${a.assignee.organization_id}"
      )
      job_type  = a.job_type
      principal = a.principal
    }
  ]
  assignment_keys = [for a in local.assignment_list : "${a.assignee}|${a.job_type}|${a.principal}"]
  assignments     = zipmap(local.assignment_keys, local.assignment_list)

  # The same planton-ai_* label set the Pulumi module applies. User labels
  # merge in first so the platform attribution labels can never be
  # clobbered by a spec label with the same key.
  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = var.metadata.name
    "planton-ai_kind"     = "gcpbigqueryreservation"
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
