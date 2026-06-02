terraform {
  required_version = ">= 1.0"

  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.0"
    }
    # Pinned to the v3 line: main.tf uses the helm-provider-v3 `set = [{...}]` attribute
    # syntax (v2 nested `set {}` blocks are rejected by v3). Keep this and the syntax aligned.
    helm = {
      source  = "hashicorp/helm"
      version = ">= 3.0"
    }
  }
}

# The kubernetes and helm providers are configured by the calling module/workspace
# with the target cluster's credentials.
provider "kubernetes" {
}
