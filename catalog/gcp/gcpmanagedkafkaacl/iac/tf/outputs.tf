output "name" {
  description = "Full resource name of the ACL (projects/{project}/locations/{location}/clusters/{cluster}/acls/{acl_id})"
  value       = google_managed_kafka_acl.this.name
}

output "resource_type" {
  description = "The resource type Google derived from acl_id: CLUSTER, TOPIC, GROUP, or TRANSACTIONAL_ID"
  value       = google_managed_kafka_acl.this.resource_type
}

output "resource_name" {
  description = "The resource name Google derived from acl_id (kafka-cluster for the cluster pattern; may be the wildcard *)"
  value       = google_managed_kafka_acl.this.resource_name
}

output "pattern_type" {
  description = "LITERAL or PREFIXED"
  value       = google_managed_kafka_acl.this.pattern_type
}
