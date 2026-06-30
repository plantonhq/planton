output "instance_id" {
  description = "OCID of the compute instance"
  value       = oci_core_instance.this.id
}

output "private_ip" {
  description = "Private IP address of the primary VNIC"
  value       = oci_core_instance.this.private_ip
}

output "public_ip" {
  description = "Public IP address of the primary VNIC"
  value       = oci_core_instance.this.public_ip
}

output "boot_volume_id" {
  description = "OCID of the boot volume"
  value       = oci_core_instance.this.boot_volume_id
}

output "availability_domain" {
  description = "Availability domain where the instance was placed"
  value       = oci_core_instance.this.availability_domain
}
