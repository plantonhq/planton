# ── Core Outputs ─────────────────────────────────────────────────────────────

output "instance_id" {
  description = "The unique identifier of the RDB instance (regional format)"
  value       = scaleway_rdb_instance.instance.id
}

# ── Public Endpoint ──────────────────────────────────────────────────────────

output "endpoint_ip" {
  description = "Public endpoint IP address"
  value       = scaleway_rdb_instance.instance.endpoint_ip
}

output "endpoint_port" {
  description = "Public endpoint port number"
  value       = scaleway_rdb_instance.instance.endpoint_port
}

# ── Private Network Endpoint ─────────────────────────────────────────────────

output "private_endpoint_ip" {
  description = "Private Network endpoint IP (empty if no PN attached)"
  value       = local.has_private_network ? try(scaleway_rdb_instance.instance.private_network[0].ip, "") : ""
}

output "private_endpoint_port" {
  description = "Private Network endpoint port (0 if no PN attached)"
  value       = local.has_private_network ? try(scaleway_rdb_instance.instance.private_network[0].port, 0) : 0
}

# ── Security ─────────────────────────────────────────────────────────────────

output "certificate" {
  description = "TLS certificate in PEM format for verifying the database server"
  value       = scaleway_rdb_instance.instance.certificate
}
