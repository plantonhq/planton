# Outputs — identical names and derivations in the Pulumi module
# (KubernetesTektonOperatorOutputs).

output "namespace" {
  description = "Namespace the operator is installed into (always tekton-operator — the release manifest's fixed namespace)"
  value       = local.namespace
}

# The images the cluster pulls, read from the image table whether or not
# image_registry is set (Pulumi twin: buildOutputs). The per-build images are
# the ones Tekton injects into every TaskRun pod; a cluster that cannot pull
# one fails every build.
locals {
  output_image_registry = local.image_registry != "" ? local.image_registry : local.upstream_registry
  output_images         = { for i in local.image_table.images : i.name => i.image }
  output_image = {
    for name, image in local.output_images : name => (
      local.image_registry != "" && startswith(image, "${local.upstream_registry}/")
      ? "${local.image_registry}${trimprefix(image, local.upstream_registry)}"
      : image
    )
  }
}

output "image_registry" {
  description = "Registry every Tekton component image is pulled from: spec.image_registry when set, else ghcr.io"
  value       = local.output_image_registry
}

output "entrypoint_image" {
  description = "Entrypoint image as the cluster pulls it, with its digest -- injected into every TaskRun pod"
  value       = local.output_image["IMAGE_PIPELINES_ARG__ENTRYPOINT_IMAGE"]
}

output "nop_image" {
  description = "Nop image as the cluster pulls it, with its digest -- stops sidecars in every TaskRun pod"
  value       = local.output_image["IMAGE_PIPELINES_ARG__NOP_IMAGE"]
}

output "workingdirinit_image" {
  description = "Workingdirinit image as the cluster pulls it, with its digest -- prepares a TaskRun's working directories"
  value       = local.output_image["IMAGE_PIPELINES_ARG__WORKINGDIRINIT_IMAGE"]
}

output "sidecarlogresults_image" {
  description = "Sidecarlogresults image as the cluster pulls it, with its digest -- carries results through sidecar logs"
  value       = local.output_image["IMAGE_PIPELINES_ARG__SIDECARLOGRESULTS_IMAGE"]
}
