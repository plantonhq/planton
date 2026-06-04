##############################################
# provider.tf
#
# Provider configuration for Kubernetes and Helm.
##############################################

terraform {
  required_version = ">= 1.0"

  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.20"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 3.0"
    }
  }
}

# Provider configuration is expected to be passed from the calling module
# or configured via environment variables (KUBECONFIG, etc.)

