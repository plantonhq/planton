output "instance_id" {
  description = "The Redis instance ID"
  value       = alicloud_kvstore_instance.main.id
}

output "connection_domain" {
  description = "The intranet connection domain"
  value       = alicloud_kvstore_instance.main.connection_domain
}

output "private_connection_port" {
  description = "The private connection port"
  value       = alicloud_kvstore_instance.main.private_connection_port
}

output "private_ip" {
  description = "The private IP address"
  value       = alicloud_kvstore_instance.main.private_ip
}
