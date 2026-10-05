# Outputs — flattened onto KubernetesSolrOperatorOutputs by the
# platform. Keep in lockstep with the Pulumi module's exports.

output "namespace" {
  description = "Kubernetes namespace the operator was installed into"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name of the operator (metadata.name)"
  value       = local.release_name
}

output "deployment_name" {
  description = "Name of the Solr operator Deployment (the chart's fullname template replayed)"
  value       = local.deployment_name
}
