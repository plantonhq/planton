output "namespace" {
  description = "The namespace in which MongoDB is deployed."
  value       = local.namespace
}

output "service" {
  description = "The base name of the MongoDB service (Helm fullname override)."
  value       = local.kube_service_name
}

output "kube_endpoint" {
  description = "Internal DNS name of the MongoDB service within the cluster."
  value       = "${local.kube_service_name}.${local.namespace}.svc.cluster.local"
}

output "port_forward_command" {
  description = "Handy command to port-forward traffic to the MongoDB service on localhost:8080 -> 27017."
  value       = "kubectl port-forward -n ${local.namespace} service/${local.kube_service_name} 8080:27017"
}

output "external_hostname" {
  description = "The external hostname for MongoDB if ingress is enabled."
  value       = local.ingress_external_hostname
}

output "username" {
  description = "The default MongoDB root username."
  value       = "root"
}

output "password_secret_name" {
  description = "Name of the Secret holding the MongoDB root password."
  value       = local.password_secret_name
}

output "password_secret_key" {
  description = "Key within the Secret that contains the MongoDB root password."
  value       = "mongodb-root-password"
}
