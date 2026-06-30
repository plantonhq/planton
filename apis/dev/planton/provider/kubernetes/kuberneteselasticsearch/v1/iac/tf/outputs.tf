###########################
# outputs.tf
###########################

output "namespace" {
  description = "The namespace where Elasticsearch and Kibana resources are created."
  value       = local.namespace
}

output "elasticsearch_service" {
  description = "The name of the Kubernetes service for Elasticsearch."
  value       = local.elasticsearch_kube_service_name
}

output "elasticsearch_port_forward_command" {
  description = "Handy port-forward command for Elasticsearch."
  value       = local.elasticsearch_kube_port_forward_command
}

output "elasticsearch_kube_endpoint" {
  description = "FQDN of the Elasticsearch service within the cluster."
  value       = local.elasticsearch_kube_service_fqdn
}

output "elasticsearch_external_hostname" {
  description = "Elasticsearch external ingress hostname (if ingress is enabled)."
  value       = local.elasticsearch_ingress_external_hostname
}

output "elasticsearch_username" {
  description = "Elasticsearch username (default 'elastic')."
  value       = "elastic"
}

output "elasticsearch_password_secret_name" {
  description = "Name of the Kubernetes secret that stores the 'elastic' user password."
  value       = "${var.metadata.name}-es-elastic-user"
}

output "elasticsearch_password_secret_key" {
  description = "Key in the secret that stores the elastic user password."
  value       = "elastic"
}

output "kibana_service" {
  description = "The name of the Kubernetes service for Kibana."
  value       = local.kibana_kube_service_name
}

output "kibana_port_forward_command" {
  description = "Handy port-forward command for Kibana."
  value       = local.kibana_kube_port_forward_command
}

output "kibana_kube_endpoint" {
  description = "FQDN of the Kibana service within the cluster."
  value       = local.kibana_kube_service_fqdn
}

output "kibana_external_hostname" {
  description = "Kibana external ingress hostname (if ingress is enabled)."
  value       = local.kibana_ingress_external_hostname
}
