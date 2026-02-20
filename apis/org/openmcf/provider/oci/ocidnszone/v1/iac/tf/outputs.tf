output "zone_id" {
  description = "OCID of the DNS zone"
  value       = oci_dns_zone.this.id
}

output "nameservers" {
  description = "Comma-separated list of authoritative nameserver hostnames"
  value       = join(",", [for ns in oci_dns_zone.this.nameservers : ns.hostname])
}
