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
#  - ../../e2e/manifest.yaml for usage examples
#  - ../README.md for kind's documentation
#  - ../../GUIDE.md for deployment patterns
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
# - Standard (GA/beta): GatewayClass, Gateway, ListenerSet, HTTPRoute,
#   GRPCRoute, TLSRoute, TCPRoute, UDPRoute, ReferenceGrant, BackendTLSPolicy
# - Experimental: everything in standard (with experimental fields) plus the
#   x- experimental resources
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
