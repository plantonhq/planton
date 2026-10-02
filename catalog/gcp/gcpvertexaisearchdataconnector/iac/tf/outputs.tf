output "name" {
  description = "Full resource name of the connector (projects/{project}/locations/{location}/collections/{collection_id}/dataConnector)"
  value       = google_discovery_engine_data_connector.this.name
}

output "collection_id" {
  description = "The collection's id -- what an engine's collection_id names"
  value       = google_discovery_engine_data_connector.this.collection_id
}

output "location" {
  description = "The collection's location (global, us, or eu)"
  value       = google_discovery_engine_data_connector.this.location
}

output "state" {
  description = "The connector's state (CREATING, ACTIVE, RUNNING, WARNING, FAILED, ...)"
  value       = google_discovery_engine_data_connector.this.state
}

# One data store per entity, in manifest order; Google fills each entity's
# data_store after setup.
output "entity_data_stores" {
  description = "Full resource names of the data stores Google created, one per entity, in manifest order"
  value       = [for entity in google_discovery_engine_data_connector.this.entities : entity.data_store if entity.data_store != null && entity.data_store != ""]
}

output "static_ip_addresses" {
  description = "The static egress IP addresses when static_ip_enabled"
  value       = google_discovery_engine_data_connector.this.static_ip_addresses
}

output "private_connectivity_project_id" {
  description = "The tenant project behind a private-connectivity connector"
  value       = google_discovery_engine_data_connector.this.private_connectivity_project_id
}
