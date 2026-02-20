output "bastion_id" {
  description = "OCID of the bastion"
  value       = oci_bastion_bastion.this.id
}

output "private_endpoint_ip_address" {
  description = "Private IP address of the bastion's endpoint in the target subnet"
  value       = oci_bastion_bastion.this.private_endpoint_ip_address
}
