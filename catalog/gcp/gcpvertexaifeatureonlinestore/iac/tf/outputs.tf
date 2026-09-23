output "name" {
  description = "Full resource name of the online store (projects/{project}/locations/{location}/featureOnlineStores/{feature_online_store_id})"
  value       = google_vertex_ai_feature_online_store.this.id
}

output "feature_online_store_id" {
  description = "The online store's id"
  value       = google_vertex_ai_feature_online_store.this.name
}

output "location" {
  description = "The location the online store lives in"
  value       = google_vertex_ai_feature_online_store.this.region
}

output "public_endpoint_domain_name" {
  description = "The dedicated endpoint's public domain name (Optimized stores); empty otherwise"
  value       = try(google_vertex_ai_feature_online_store.this.dedicated_serving_endpoint[0].public_endpoint_domain_name, "")
}

output "service_attachment" {
  description = "The Private Service Connect service attachment consumers target, once PSC is enabled and a view has synced; empty otherwise"
  value       = try(google_vertex_ai_feature_online_store.this.dedicated_serving_endpoint[0].service_attachment, "")
}

# Manifest order, so the list reads the same on both engines.
output "feature_view_names" {
  description = "Full resource names of the declared feature views, in manifest order"
  value       = [for view in var.spec.feature_views : google_vertex_ai_feature_online_store_featureview.this[view.feature_view_id].id]
}
