output "name" {
  description = "Full resource name of the feature group (projects/{project}/locations/{location}/featureGroups/{feature_group_id})"
  value       = google_vertex_ai_feature_group.this.id
}

output "feature_group_id" {
  description = "The feature group's id -- what an online store's feature view references"
  value       = google_vertex_ai_feature_group.this.name
}

output "location" {
  description = "The location the feature group lives in"
  value       = google_vertex_ai_feature_group.this.region
}

# Manifest order, so the list reads the same on both engines.
output "feature_names" {
  description = "Full resource names of the registered features, in manifest order"
  value       = [for feature in var.spec.features : google_vertex_ai_feature_group_feature.this[feature.feature_id].id]
}
