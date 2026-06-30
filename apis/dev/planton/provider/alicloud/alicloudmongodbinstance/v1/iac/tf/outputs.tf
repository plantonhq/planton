output "instance_id" {
  description = "The MongoDB instance ID assigned by Alibaba Cloud"
  value       = alicloud_mongodb_instance.main.id
}

output "replica_set_name" {
  description = "The replica set name for use in MongoDB connection strings"
  value       = alicloud_mongodb_instance.main.replica_set_name
}
