output "instance_id" {
  description = "The RocketMQ instance ID"
  value       = alicloud_rocketmq_instance.main.id
}

output "tcp_endpoint" {
  description = "TCP endpoint for VPC-internal access"
  value = try(
    [for ep in alicloud_rocketmq_instance.main.network_info[0].endpoints :
      ep.endpoint_url if ep.endpoint_type == "TCP_VPC"
    ][0],
    ""
  )
}

output "internet_endpoint" {
  description = "TCP endpoint for public internet access"
  value = try(
    [for ep in alicloud_rocketmq_instance.main.network_info[0].endpoints :
      ep.endpoint_url if ep.endpoint_type == "TCP_INTERNET"
    ][0],
    ""
  )
}

output "topic_ids" {
  description = "Map of topic names to their resource IDs"
  value = {
    for name, t in alicloud_rocketmq_topic.topics : name => t.id
  }
}

output "consumer_group_ids" {
  description = "Map of consumer group IDs to their resource IDs"
  value = {
    for id, cg in alicloud_rocketmq_consumer_group.consumer_groups : id => cg.id
  }
}
