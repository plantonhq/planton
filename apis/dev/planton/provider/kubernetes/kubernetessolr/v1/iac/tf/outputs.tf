output "namespace" {
  description = "The namespace where Solr and related resources are created."
  value       = local.namespace
}

output "service" {
  description = "The Kubernetes service name for the SolrCloud common service."
  value       = local.kube_service_name
}

output "kube_endpoint" {
  description = "The internal service FQDN for the SolrCloud service."
  value       = local.kube_service_fqdn
}

output "port_forward_command" {
  description = "A handy command to port-forward local 8080 to the Solr service 8080."
  value       = local.kube_port_forward_command
}

output "external_hostname" {
  description = "The external Istio gateway hostname, if ingress is enabled."
  value       = local.ingress_external_hostname
}

output "internal_hostname" {
  description = "The internal Istio gateway hostname, if ingress is enabled."
  value       = local.ingress_internal_hostname
}
