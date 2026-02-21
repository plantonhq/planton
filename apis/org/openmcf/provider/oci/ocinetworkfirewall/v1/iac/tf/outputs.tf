output "firewall_id" {
  description = "OCID of the network firewall"
  value       = oci_network_firewall_network_firewall.this.id
}

output "ipv4_address" {
  description = "IPv4 address of the firewall appliance"
  value       = oci_network_firewall_network_firewall.this.ipv4address
}

output "policy_id" {
  description = "OCID of the firewall policy"
  value       = oci_network_firewall_network_firewall_policy.this.id
}
