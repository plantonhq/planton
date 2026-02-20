output "stream_pool_id" {
  description = "OCID of the stream pool"
  value       = oci_streaming_stream_pool.this.id
}

output "endpoint_fqdn" {
  description = "FQDN for accessing streams in the pool"
  value       = oci_streaming_stream_pool.this.endpoint_fqdn
}

output "kafka_bootstrap_servers" {
  description = "Kafka-compatible bootstrap server string"
  value       = try(oci_streaming_stream_pool.this.kafka_settings[0].bootstrap_servers, "")
}
