# Provider requirements for the KubernetesAgentSandboxController module.
#
# kubectl (alekc/kubectl) applies upstream's documents with server-side
# apply. http fetches the pinned release asset and the router's deployment
# files. kubernetes reads, at plan time, the Deployments the module is about
# to apply (the workload ownership check) and the definitions the kept-CRD
# block is about to write. helm is declared because helm_crds.tf is one
# generated file for chart and bundle kinds alike, and init resolves every
# data source in it; this module never renders a chart.
#
# Providers are configured by the calling workspace (the kubeconfig
# environment contract).

terraform {
  required_version = ">= 1.4"

  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.35"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 3.0"
    }
    kubectl = {
      source  = "alekc/kubectl"
      version = ">= 2.0"
    }
    http = {
      source  = "hashicorp/http"
      version = "~> 3.4"
    }
  }
}

provider "kubernetes" {
}

provider "helm" {
}

provider "kubectl" {
}
