terraform {
  required_version = ">= 1.0"

  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.0"
    }
    # Pinned to the v3 major: main.tf uses the helm-provider-v3 values/`set = [{...}]`
    # attribute syntax (v2 nested `set {}` blocks are rejected by v3). The ~> 3.0 pin keeps
    # the major fixed so a future helm v4 cannot silently reintroduce a breaking change.
    helm = {
      source  = "hashicorp/helm"
      version = "~> 3.0"
    }
  }
}

# The kubernetes and helm providers are configured by the calling module/workspace
# with the target cluster's credentials.
provider "kubernetes" {
}
