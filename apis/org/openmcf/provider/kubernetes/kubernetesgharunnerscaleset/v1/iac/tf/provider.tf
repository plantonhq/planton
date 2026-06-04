##############################################
# provider.tf
#
# Terraform provider configuration for KubernetesGhaRunnerScaleSet
##############################################

terraform {
  required_version = ">= 1.0"

  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 3.0"
    }
  }
}

# Kubernetes and Helm providers should be configured by the caller
# using kubeconfig or in-cluster configuration.

