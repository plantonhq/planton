output "name" {
  description = "Full resource name of the collection (projects/{project}/locations/{location}/collections/{collection_id})"
  value       = google_vector_search_collection.this.name
}

output "collection_id" {
  description = "The collection's ID"
  value       = google_vector_search_collection.this.collection_id
}

output "location" {
  description = "The collection's location"
  value       = google_vector_search_collection.this.location
}

# Manifest order, so the list reads the same on both engines.
output "index_names" {
  description = "Full resource names of the indexes declared on the collection, in manifest order"
  value       = [for index in var.spec.indexes : google_vector_search_index.this[index.index_id].name]
}

output "index_count" {
  description = "Number of indexes declared on the collection"
  value       = length(var.spec.indexes)
}
