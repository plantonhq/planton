output "name" {
  description = "Full resource name of the feature (projects/{project}/locations/{location}/features/{feature})"
  value       = google_gke_hub_feature.this.id
}
