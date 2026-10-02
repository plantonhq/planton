locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # Google requires a display name; the spec defaults it to metadata.name --
  # identical to the Pulumi module.
  display_name = var.spec.display_name != "" ? var.spec.display_name : var.metadata.name

  # Empty optional strings become null so the provider omits them from the
  # API payload instead of sending empty values it would reject or diff on.
  description     = var.spec.description != "" ? var.spec.description : null
  kms_key_name    = var.spec.kms_key_name != "" ? var.spec.kms_key_name : null
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # The same planton-ai_* label set the Pulumi module applies, on the
  # TensorBoard and on every declared experiment and run. User labels merge
  # in first so the platform attribution labels can never be clobbered by a
  # spec label with the same key.
  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = var.metadata.name
    "planton-ai_kind"     = "gcpvertexaitensorboard"
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

  # Experiments and runs address their TensorBoard by the numeric id Google
  # assigned -- the last segment of its resource name.
  tensorboard_id = reverse(split("/", google_vertex_ai_tensorboard.this.name))[0]

  # The folded experiments keyed by experiment_id, and their runs keyed by
  # "{experiment_id}/{run_id}", so adding or removing one never renumbers
  # the others.
  experiments = { for experiment in var.spec.experiments : experiment.experiment_id => experiment }
  runs = merge([
    for experiment in var.spec.experiments : {
      for run in experiment.runs : "${experiment.experiment_id}/${run.run_id}" => {
        experiment_id = experiment.experiment_id
        run_id        = run.run_id
        display_name  = run.display_name
        description   = run.description
        labels        = run.labels
      }
    }
  ]...)
}
