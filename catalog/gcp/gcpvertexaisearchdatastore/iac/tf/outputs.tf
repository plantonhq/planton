output "name" {
  description = "Full resource name of the data store (projects/{project}/locations/{location}/collections/default_collection/dataStores/{data_store_id})"
  value       = google_discovery_engine_data_store.this.name
}

output "data_store_id" {
  description = "The data store's id -- what an engine's data_store_ids lists"
  value       = google_discovery_engine_data_store.this.data_store_id
}

output "location" {
  description = "The data store's location (global, us, or eu)"
  value       = google_discovery_engine_data_store.this.location
}

# The store's default schema: the declared schema's id when one is declared
# (Google makes it the default once it exists, after the store is created,
# so the store's own attribute reads empty at apply and flips on the first
# refresh -- measured), otherwise the schema Google created with the store.
output "default_schema_id" {
  description = "The id of the store's default schema"
  value       = var.spec.schema != null ? var.spec.schema.schema_id : google_discovery_engine_data_store.this.default_schema_id
}

output "schema_name" {
  description = "Full resource name of the declared schema (empty when none)"
  value       = var.spec.schema != null ? one(google_discovery_engine_schema.this[*].name) : ""
}

# Manifest order, so the lists read the same on both engines.
output "target_site_names" {
  description = "Full resource names of the target sites, in manifest order"
  value       = [for site in var.spec.target_sites : google_discovery_engine_target_site.this[site.provided_uri_pattern].name]
}

output "sitemap_names" {
  description = "Full resource names of the sitemaps, in manifest order"
  value       = [for uri in var.spec.sitemap_uris : google_discovery_engine_sitemap.this[uri].name]
}
