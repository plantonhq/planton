output "load_balancer_id" {
  description = "OCID of the load balancer."
  value       = oci_load_balancer_load_balancer.this.id
}

output "ip_addresses" {
  description = "Comma-separated IP addresses assigned to the load balancer."
  value       = join(",", [for ip in oci_load_balancer_load_balancer.this.ip_address_details : ip.ip_address])
}
