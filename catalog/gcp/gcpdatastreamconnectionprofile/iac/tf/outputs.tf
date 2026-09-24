output "name" {
  description = "Full resource name of the profile (projects/{project}/locations/{location}/connectionProfiles/{connection_profile_id}) -- what streams reference as their source or destination"
  value       = google_datastream_connection_profile.this.name
}

output "connection_profile_id" {
  description = "The profile's id"
  value       = google_datastream_connection_profile.this.connection_profile_id
}
