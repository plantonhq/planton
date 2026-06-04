output "namespace" {
  description = "The namespace in which the Postgres resources are deployed."
  value       = local.namespace
}

output "service" {
  description = "Name of the Postgres service (master)."
  value       = local.kube_service_name
}

output "port_forward_command" {
  description = "Convenient command to port-forward to the Postgres service."
  value       = local.kube_port_forward_command
}

output "kube_endpoint" {
  description = "FQDN of the Postgres service within the cluster."
  value       = local.kube_service_fqdn
}

output "external_hostname" {
  description = "The external hostname for Postgres if ingress is enabled."
  value       = local.ingress_external_hostname
}

# Nested objects so the generic outputs transformer (pkg/outputs.Flatten) produces
# the dotted keys password_secret.name / password_secret.key that the
# KubernetesPostgresStackOutputs proto's password_secret (KubernetesSecretKey)
# field expects -- matching the Pulumi module's "password_secret.name" exports.
# The secret name follows the Zalando operator convention for the superuser
# credentials of the "db-<metadata.name>" cluster.
output "password_secret" {
  description = "Kubernetes secret key for the Postgres superuser password."
  value = {
    name = "postgres.db-${var.metadata.name}.credentials.postgresql.acid.zalan.do"
    key  = "password"
  }
}

output "username_secret" {
  description = "Kubernetes secret key for the Postgres superuser username."
  value = {
    name = "postgres.db-${var.metadata.name}.credentials.postgresql.acid.zalan.do"
    key  = "username"
  }
}
