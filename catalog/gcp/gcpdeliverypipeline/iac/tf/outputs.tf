output "name" {
  description = "Full resource name of the pipeline (projects/{project}/locations/{location}/deliveryPipelines/{delivery_pipeline_id})"
  value       = google_clouddeploy_delivery_pipeline.this.id
}

output "delivery_pipeline_id" {
  description = "The pipeline's ID"
  value       = google_clouddeploy_delivery_pipeline.this.name
}

output "uid" {
  description = "Google's unique identifier for the pipeline"
  value       = google_clouddeploy_delivery_pipeline.this.uid
}
