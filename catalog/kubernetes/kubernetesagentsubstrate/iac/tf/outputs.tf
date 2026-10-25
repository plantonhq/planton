# Outputs (outputs.go): exactly what a product that creates actors on this
# control plane reads.

output "namespace" {
  value = local.namespace
}

output "api_address" {
  value = "${local.api_server_host}:443"
}

output "api_server_name" {
  value = local.api_server_host
}

output "api_token_audience" {
  value = local.api_server_host
}

output "api_trust_bundle_name" {
  value = local.api_trust_bundle_name
}

output "api_trust_bundle_signer" {
  value = local.api_trust_bundle_sign
}

output "api_access_namespace_label" {
  value = "${local.api_access_label_key}=${local.api_access_label_value}"
}

# Only agentgateway still serves plaintext listeners; the Envoy router serves
# only TLS (outputs.go plaintextRouterEndpoint).
output "router_endpoint" {
  value = local.router.agentgateway ? "http://${local.router_name}.${local.namespace}.svc:80" : ""
}

output "router_tls_endpoint" {
  value = "https://${local.router_name}.${local.namespace}.svc:443"
}

output "storage_location" {
  value = local.storage_location
}

output "actor_jwt_issuer" {
  value = local.api.actor_issuer
}

output "gvisor_worker_image" {
  value = try(local.mirrored[local.images["ateworker-gvisor"]], "")
}

output "microvm_worker_image" {
  value = try(local.mirrored[local.images["ateworker-microvm"]], "")
}

output "version" {
  value = local.version
}

output "router_client_signer_name" {
  value = local.router_client_signer
}
