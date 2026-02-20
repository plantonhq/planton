output "network_security_group_id" {
  description = "OCID of the network security group"
  value       = oci_core_network_security_group.this.id
}
