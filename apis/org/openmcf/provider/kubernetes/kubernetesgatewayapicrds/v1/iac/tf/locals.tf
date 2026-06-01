##############################################
# locals.tf
#
# Computed local values for the
# KubernetesGatewayApiCrds module.
##############################################

locals {
  # Gateway API version
  version = coalesce(var.spec.version, "v1.2.1")

  # Determine if experimental channel is requested
  is_experimental = try(var.spec.install_channel.channel, "standard") == "experimental"

  # Channel name for outputs
  channel_name = local.is_experimental ? "experimental" : "standard"

  # Manifest filename based on channel
  manifest_file = local.is_experimental ? "experimental-install.yaml" : "standard-install.yaml"

  # Full URL to download CRD manifests
  manifest_url = "https://github.com/kubernetes-sigs/gateway-api/releases/download/${local.version}/${local.manifest_file}"

  # Resource labels
  labels = {
    "app.kubernetes.io/name"       = "gateway-api-crds"
    "app.kubernetes.io/instance"   = var.metadata.name
    "app.kubernetes.io/managed-by" = "openmcf"
    "app.kubernetes.io/component"  = "crds"
    "gateway-api/version"          = local.version
    "gateway-api/channel"          = local.channel_name
  }
}
