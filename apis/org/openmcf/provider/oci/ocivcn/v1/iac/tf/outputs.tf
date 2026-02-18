output "vcn_id" {
  description = "OCID of the VCN"
  value       = oci_core_vcn.this.id
}

output "default_route_table_id" {
  description = "OCID of the default route table created with the VCN"
  value       = oci_core_vcn.this.default_route_table_id
}

output "default_security_list_id" {
  description = "OCID of the default security list created with the VCN"
  value       = oci_core_vcn.this.default_security_list_id
}

output "default_dhcp_options_id" {
  description = "OCID of the default DHCP options created with the VCN"
  value       = oci_core_vcn.this.default_dhcp_options_id
}

output "internet_gateway_id" {
  description = "OCID of the Internet Gateway (empty when not enabled)"
  value       = var.spec.is_internet_gateway_enabled ? oci_core_internet_gateway.this[0].id : ""
}

output "nat_gateway_id" {
  description = "OCID of the NAT Gateway (empty when not enabled)"
  value       = var.spec.is_nat_gateway_enabled ? oci_core_nat_gateway.this[0].id : ""
}

output "service_gateway_id" {
  description = "OCID of the Service Gateway (empty when not enabled)"
  value       = var.spec.is_service_gateway_enabled ? oci_core_service_gateway.this[0].id : ""
}
