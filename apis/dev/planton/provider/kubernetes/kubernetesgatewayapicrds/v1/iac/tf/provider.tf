##############################################
# provider.tf
#
# Provider configuration for the
# KubernetesGatewayApiCrds module.
#
# The Kubernetes provider is expected to be
# configured by the calling module/workspace.
##############################################

terraform {
  required_version = ">= 1.0"

  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.0"
    }
    http = {
      source  = "hashicorp/http"
      version = ">= 3.0"
    }
    kubectl = {
      source  = "alekc/kubectl"
      version = ">= 2.0"
    }
  }
}

# Note: The kubernetes and kubectl providers must be configured
# by the calling module with appropriate credentials.
#
# Example:
# provider "kubernetes" {
#   config_path = "~/.kube/config"
# }
#
# provider "kubectl" {
#   config_path = "~/.kube/config"
# }
