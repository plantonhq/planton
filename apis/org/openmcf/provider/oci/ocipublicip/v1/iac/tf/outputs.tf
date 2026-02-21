output "public_ip_id" {
  description = "OCID of the created public IP"
  value       = oci_core_public_ip.this.id
}

output "ip_address" {
  description = "The allocated IPv4 address"
  value       = oci_core_public_ip.this.ip_address
}
