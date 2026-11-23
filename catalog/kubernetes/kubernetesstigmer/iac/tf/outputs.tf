# Every output derives from the declaration: the operator names everything
# after the Stigmer, so none needs a read of the cluster. Pulumi twin:
# module/outputs.go.

output "namespace" {
  value = local.namespace
}

output "operator_namespace" {
  value = local.operator_namespace
}

output "operator_release" {
  value = local.operator_release
}

output "server_endpoint" {
  value = "http://${local.stigmer_name}.${local.namespace}.svc.cluster.local:${local.server_port}"
}

output "artifacts_endpoint" {
  value = "http://${local.stigmer_name}.${local.namespace}.svc.cluster.local:${local.artifacts_port}"
}

output "url" {
  value = try(var.spec.ingress.hostname, "") != "" ? "https://${var.spec.ingress.hostname}" : "http://localhost:${local.server_port}"
}

output "port_forward_command" {
  value = "kubectl -n ${local.namespace} port-forward svc/${local.stigmer_name} ${local.server_port}:${local.server_port} ${local.artifacts_port}:${local.artifacts_port}"
}

output "status_command" {
  value = "kubectl -n ${local.namespace} get stigmer ${local.stigmer_name}"
}

output "owner_password_command" {
  value = local.bundled_sign_in ? "kubectl -n ${local.namespace} get secret ${local.stigmer_name}-owner-setup -o jsonpath='{.data.password}' | base64 -d" : ""
}

output "keys_secret" {
  value = local.keys_secret
}

# The keys Secret as a manifest that carries only its name and data, so
# `kubectl -n <namespace> apply -f stigmer-keys.yaml` recreates it in any
# cluster.
output "keys_export_command" {
  value = "kubectl -n ${local.namespace} get secret ${local.keys_secret} -o go-template='{{\"apiVersion: v1\\nkind: Secret\\nmetadata:\\n  name: ${local.keys_secret}\\ntype: Opaque\\ndata:\\n\"}}{{range $k, $v := .data}}{{printf \"  %s: %s\\n\" $k $v}}{{end}}' > ${local.keys_secret}.yaml"
}
