# ── Core Outputs ─────────────────────────────────────────────────────────────

output "instance_id" {
  description = "The unique identifier of the MongoDB instance (regional format)"
  value       = scaleway_mongodb_instance.instance.id
}

# ── Public Network Endpoint ──────────────────────────────────────────────────

output "public_dns_record" {
  description = "Public endpoint DNS record (empty if private-only)"
  value       = try(scaleway_mongodb_instance.instance.public_network[0].dns_record, "")
}

output "public_port" {
  description = "Public endpoint port number (0 if private-only)"
  value       = try(scaleway_mongodb_instance.instance.public_network[0].port, 0)
}

# ── Private Network Endpoint ─────────────────────────────────────────────────

output "private_dns_records" {
  description = "Private Network endpoint DNS records (empty if no PN attached)"
  value       = local.has_private_network ? try(scaleway_mongodb_instance.instance.private_network[0].dns_records, []) : []
}

output "private_ips" {
  description = "Private Network endpoint IP addresses (empty if no PN attached)"
  value       = local.has_private_network ? try(scaleway_mongodb_instance.instance.private_network[0].ips, []) : []
}

output "private_port" {
  description = "Private Network endpoint port (0 if no PN attached)"
  value       = local.has_private_network ? try(scaleway_mongodb_instance.instance.private_network[0].port, 0) : 0
}

# ── Security ─────────────────────────────────────────────────────────────────

output "tls_certificate" {
  description = "TLS certificate in PEM format for verifying the database server"
  value       = scaleway_mongodb_instance.instance.tls_certificate
}
