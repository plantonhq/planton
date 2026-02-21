output "instance_id" {
  description = "The ECS instance ID assigned by Alibaba Cloud"
  value       = alicloud_instance.main.id
}

output "private_ip" {
  description = "The primary private IP address of the ECS instance"
  value       = alicloud_instance.main.private_ip
}

output "public_ip" {
  description = "The public IP address of the ECS instance (empty if no public IP)"
  value       = alicloud_instance.main.public_ip
}
