# Provider requirements for the KubernetesTektonOperator module.
#
# kubectl (alekc/kubectl) applies the release manifest per document with
# server-side apply: unlike the hashicorp kubernetes provider's
# kubernetes_manifest resource, kubectl_manifest needs no cluster
# connection at plan time (offline plan proofs, single-run infra charts).
# The http provider fetches the released manifest from the pinned tag.
#
# Providers are configured by the calling workspace/environment (the same
# kubeconfig environment contract).

terraform {
  required_version = ">= 1.3" # startswith() and lifecycle preconditions (image_registry)

  required_providers {
    kubectl = {
      source  = "alekc/kubectl"
      version = ">= 2.0"
    }
    http = {
      source  = "hashicorp/http"
      version = ">= 3.0"
    }
  }
}

provider "kubectl" {
}
