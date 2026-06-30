# Server ID
output "server_id" {
  description = "The unique identifier of the created instance server"
  value       = scaleway_instance_server.server.id
}

# Public IP Address (empty string if no public IP)
output "public_ip_address" {
  description = "The public IPv4 address assigned to the instance"
  value       = length(scaleway_instance_ip.ip) > 0 ? scaleway_instance_ip.ip[0].address : ""
}

# Public IP ID (empty string if no public IP)
output "public_ip_id" {
  description = "The unique identifier of the Flexible IP resource"
  value       = length(scaleway_instance_ip.ip) > 0 ? scaleway_instance_ip.ip[0].id : ""
}

# Private IP Address (first private IP if available)
output "private_ip_address" {
  description = "The private IP address on the attached Private Network"
  value       = length(scaleway_instance_server.server.private_ips) > 0 ? scaleway_instance_server.server.private_ips[0].address : ""
}

# Instance State
output "instance_state" {
  description = "The current operational state of the instance"
  value       = scaleway_instance_server.server.state
}

# Zone
output "zone" {
  description = "The zone where the instance is deployed"
  value       = scaleway_instance_server.server.zone
}
