output "subnet_id" {
  description = "OCID of the subnet"
  value       = oci_core_subnet.this.id
}

output "subnet_domain_name" {
  description = "Fully qualified domain name of the subnet"
  value       = oci_core_subnet.this.subnet_domain_name
}

output "virtual_router_ip" {
  description = "IP address of the virtual router in this subnet"
  value       = oci_core_subnet.this.virtual_router_ip
}

output "virtual_router_mac" {
  description = "MAC address of the virtual router in this subnet"
  value       = oci_core_subnet.this.virtual_router_mac
}

output "route_table_id" {
  description = "OCID of the route table associated with this subnet"
  value       = oci_core_subnet.this.route_table_id
}
