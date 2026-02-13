# ScalewayContainerRegistry Terraform Module
#
# Creates a Scaleway Container Registry namespace -- an OCI-compliant
# registry for storing and distributing container images and Helm charts.
#
# Container Registry namespaces are REGIONAL resources.
# Available regions: fr-par, nl-ams, pl-waw
#
# This module wraps a single `scaleway_registry_namespace` resource.
#
# NOTE: Scaleway Container Registry namespaces do not support tags.
# Unlike most other Scaleway resources, the registry API does not
# accept tags/labels. Standard OpenMCF metadata tags are not applied.

resource "scaleway_registry_namespace" "registry" {
  name        = local.namespace_name
  description = local.description
  is_public   = local.is_public
  region      = local.region
}
