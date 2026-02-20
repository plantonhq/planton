output "gateway_id" {
  description = "OCID of the API gateway"
  value       = oci_apigateway_gateway.this.id
}

output "hostname" {
  description = "Hostname assigned to the gateway by OCI"
  value       = oci_apigateway_gateway.this.hostname
}

output "deployment_endpoint" {
  description = "Full endpoint URL of the API deployment"
  value       = oci_apigateway_deployment.this.endpoint
}
