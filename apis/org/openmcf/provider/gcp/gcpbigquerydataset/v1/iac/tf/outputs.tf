output "dataset_id" {
  description = "The short dataset ID"
  value       = google_bigquery_dataset.this.dataset_id
}

output "self_link" {
  description = "The fully qualified URI of the dataset"
  value       = google_bigquery_dataset.this.self_link
}

output "project" {
  description = "The GCP project that contains this dataset"
  value       = google_bigquery_dataset.this.project
}

output "creation_time" {
  description = "The creation time of the dataset in milliseconds since epoch"
  value       = google_bigquery_dataset.this.creation_time
}
