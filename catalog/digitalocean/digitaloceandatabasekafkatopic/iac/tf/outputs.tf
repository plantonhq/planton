# Stack outputs — exactly the DigitalOceanDatabaseKafkaTopicStackOutputs
# contract, identical across both provisioners. The (cluster, topic name)
# pair is the topic's API identity; DigitalOcean mints no standalone id.
#
# The topic's provisioning state is deliberately not exported: an apply-time
# snapshot goes stale (creation is asynchronous), so live state belongs to
# whoever reads the API, never to the outputs contract.

output "cluster_id" {
  description = "UUID of the Kafka database cluster the topic lives in"
  value       = digitalocean_database_kafka_topic.topic.cluster_id
}

output "topic_name" {
  description = "Name of the Kafka topic (its API identity within the cluster)"
  value       = digitalocean_database_kafka_topic.topic.name
}
