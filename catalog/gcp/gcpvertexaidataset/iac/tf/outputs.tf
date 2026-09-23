output "name" {
  description = "Full resource name of the dataset (projects/{project_number}/locations/{location}/datasets/{dataset_id})"
  value       = google_vertex_ai_dataset.this.name
}

output "dataset_id" {
  description = "The numeric id Google assigned the dataset"
  value       = reverse(split("/", google_vertex_ai_dataset.this.name))[0]
}

output "location" {
  description = "The location the dataset lives in"
  value       = google_vertex_ai_dataset.this.region
}
