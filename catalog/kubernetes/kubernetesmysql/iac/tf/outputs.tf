# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports (KubernetesMysqlOutputs).

output "namespace" {
  description = "Namespace the cluster runs in"
  value       = local.namespace
}

output "cluster_name" {
  description = "Name of the PerconaXtraDBCluster resource (equals metadata.name) — every operator-created object derives from it"
  value       = local.cluster_name
}

output "primary_service" {
  description = "The WRITE Service applications connect to (`<name>-haproxy` or `<name>-proxysql`)"
  value       = local.primary_service_name
}

output "replicas_service" {
  description = "The READ Service (`<name>-haproxy-replicas`) — HAProxy with the replicas Service enabled; empty otherwise"
  value       = local.replicas_service_name
}

output "kube_endpoint" {
  description = "In-cluster endpoint of the write path — the connection host for applications in the same cluster"
  value       = local.kube_endpoint
}

output "port_forward_command" {
  description = "Port-forward command for reaching the database from a workstation when no exposure is composed"
  value       = "kubectl port-forward svc/${local.primary_service_name} -n ${local.namespace} 3306:3306"
}

output "root_password_secret" {
  description = "Secret key holding the root password (the operator-managed `<name>-secrets` system-users Secret, key \"root\")"
  value = {
    name = local.users_secret_name
    key  = "root"
  }
}
