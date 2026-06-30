output "vpn_gateway_id" {
  description = "The VPN Gateway ID"
  value       = alicloud_vpn_gateway.main.id
}

output "internet_ip" {
  description = "The VPN Gateway's public internet IP address"
  value       = alicloud_vpn_gateway.main.internet_ip
}

output "ssl_vpn_internet_ip" {
  description = "The SSL VPN internet IP address (populated when SSL VPN is enabled)"
  value       = alicloud_vpn_gateway.main.ssl_vpn_internet_ip
}

output "connection_ids" {
  description = "Map of connection name to VPN connection ID"
  value = {
    for name, conn in alicloud_vpn_connection.connections : name => conn.id
  }
}
