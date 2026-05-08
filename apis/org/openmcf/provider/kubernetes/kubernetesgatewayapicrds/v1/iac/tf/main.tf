##############################################
# main.tf
#
# Main orchestration file for installing
# Kubernetes Gateway API CRDs.
#
# This module fetches and applies the official
# Gateway API CRD manifests from the kubernetes-sigs
# gateway-api GitHub repository.
#
# Resources Created:
#  1. Gateway API CRDs (cluster-scoped)
#
# For more information see:
#  - examples.md for usage examples
#  - ../README.md for component documentation
#  - ../../docs/README.md for deployment patterns
##############################################

##############################################
# 1. Fetch Gateway API CRD Manifest
#
# Downloads the CRD manifest YAML from the
# official Gateway API GitHub releases.
##############################################
data "http" "gateway_api_crds" {
  url = local.manifest_url

  request_headers = {
    Accept = "application/yaml"
  }
}

##############################################
# 2. Apply Gateway API CRDs
#
# The Gateway API CRDs are cluster-scoped
# resources that enable Gateway, HTTPRoute,
# GRPCRoute, and other Gateway API resources.
#
# Depending on the channel:
# - Standard: Gateway, GatewayClass, HTTPRoute, ReferenceGrant
# - Experimental: Standard + TCPRoute, UDPRoute, TLSRoute, GRPCRoute
#
# Note: CRDs are applied using kubectl_manifest
# which handles multi-document YAML properly.
##############################################
resource "kubectl_manifest" "gateway_api_crds" {
  for_each = {
    for idx, doc in split("---", data.http.gateway_api_crds.response_body) : idx => doc
    if trimspace(doc) != "" && can(yamldecode(doc))
  }

  yaml_body = each.value

  server_side_apply = true
  force_conflicts   = true
}
