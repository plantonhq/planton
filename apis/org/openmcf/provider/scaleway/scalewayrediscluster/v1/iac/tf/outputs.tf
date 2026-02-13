# ── Core Outputs ─────────────────────────────────────────────────────────────

output "cluster_id" {
  description = "The unique identifier of the Redis cluster (zonal format)"
  value       = scaleway_redis_cluster.cluster.id
}

# ── Public Network Endpoint ──────────────────────────────────────────────────

output "public_network_port" {
  description = "Public network endpoint port (0 if using Private Network)"
  value       = local.has_private_network ? 0 : try(scaleway_redis_cluster.cluster.public_network[0].port, 0)
}

output "public_network_ips" {
  description = "Public network endpoint IPs (empty if using Private Network)"
  value       = local.has_private_network ? [] : try(scaleway_redis_cluster.cluster.public_network[0].ips, [])
}

# ── Private Network Endpoint ─────────────────────────────────────────────────

output "private_network_port" {
  description = "Private Network endpoint port (0 if not using Private Network)"
  value       = local.has_private_network ? try(tolist(scaleway_redis_cluster.cluster.private_network)[0].port, 0) : 0
}

output "private_network_ips" {
  description = "Private Network endpoint IPs (empty if not using Private Network)"
  value       = local.has_private_network ? try(tolist(scaleway_redis_cluster.cluster.private_network)[0].ips, []) : []
}

# ── Security ─────────────────────────────────────────────────────────────────

output "certificate" {
  description = "TLS certificate in PEM format (empty if TLS disabled)"
  value       = scaleway_redis_cluster.cluster.certificate
}
