output "name" {
  description = "Full resource name of the topic (projects/{project}/locations/{location}/clusters/{cluster}/topics/{topic_id})"
  value       = google_managed_kafka_topic.this.name
}

output "topic_id" {
  description = "The topic's name as Kafka clients use it"
  value       = google_managed_kafka_topic.this.topic_id
}
