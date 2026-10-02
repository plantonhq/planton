output "name" {
  description = "Full resource name of the stream (projects/{project}/locations/{location}/streams/{stream_id})"
  value       = google_datastream_stream.this.name
}

output "stream_id" {
  description = "The stream's id"
  value       = google_datastream_stream.this.stream_id
}
