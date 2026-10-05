# Outputs — must flatten onto KubernetesHelmReleaseOutputs
# (outputs.proto) identically to the Pulumi module's exports. Values
# come from the helm_release resource's recorded metadata, so they reflect
# what Helm actually installed.

output "namespace" {
  description = "The namespace the release is installed in"
  value       = local.namespace_name
}

output "release_name" {
  description = "The Helm release name (helm list NAME column)"
  value       = helm_release.helm_release.name
}

output "version" {
  description = "The installed chart version"
  value       = helm_release.helm_release.metadata.version
}

output "app_version" {
  description = "The chart's appVersion (the packaged application's upstream version)"
  value       = helm_release.helm_release.metadata.app_version
}

output "status" {
  description = "The release status as Helm records it (e.g. deployed)"
  value       = helm_release.helm_release.status
}

output "revision" {
  description = "The release revision number (1 on install, incremented by upgrades/rollbacks)"
  value       = helm_release.helm_release.metadata.revision
}
