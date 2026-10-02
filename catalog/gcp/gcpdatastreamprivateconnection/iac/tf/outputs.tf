output "name" {
  description = "Full resource name of the private connection (projects/{project}/locations/{location}/privateConnections/{private_connection_id}) -- what connection profiles reference"
  value       = google_datastream_private_connection.this.name
}

output "private_connection_id" {
  description = "The private connection's id"
  value       = google_datastream_private_connection.this.private_connection_id
}
