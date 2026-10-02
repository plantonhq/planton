output "name" {
  description = "Full resource name of the TensorBoard (projects/{project_number}/locations/{location}/tensorboards/{tensorboard_id}) -- what a training job's tensorboard field takes"
  value       = google_vertex_ai_tensorboard.this.name
}

output "tensorboard_id" {
  description = "The numeric id Google assigned the TensorBoard"
  value       = local.tensorboard_id
}

output "location" {
  description = "The location the TensorBoard lives in"
  value       = google_vertex_ai_tensorboard.this.region
}

output "blob_storage_path_prefix" {
  description = "Cloud Storage path prefix where the TensorBoard stores blob data"
  value       = google_vertex_ai_tensorboard.this.blob_storage_path_prefix
}

# Manifest order, so the lists read the same on both engines.
output "experiment_names" {
  description = "Full resource names of the declared experiments, in manifest order"
  value       = [for experiment in var.spec.experiments : google_vertex_ai_tensorboard_experiment.this[experiment.experiment_id].id]
}

output "run_names" {
  description = "Full resource names of the declared runs, experiment by experiment in manifest order"
  value = flatten([
    for experiment in var.spec.experiments : [
      for run in experiment.runs : google_vertex_ai_tensorboard_run.this["${experiment.experiment_id}/${run.run_id}"].id
    ]
  ])
}
