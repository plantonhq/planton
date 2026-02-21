output "network_load_balancer_id" {
  description = "OCID of the network load balancer."
  value       = oci_network_load_balancer_network_load_balancer.this.id
}

output "ip_addresses" {
  description = "Comma-separated IP addresses assigned to the network load balancer."
  value       = join(",", [for ip in oci_network_load_balancer_network_load_balancer.this.ip_addresses : ip.ip_address])
}
