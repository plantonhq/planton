# Provider requirements for the KubernetesStigmer module.
#
# helm installs the stigmer-operator chart. kubectl (alekc/kubectl) declares
# the Stigmer and the restore check's ConfigMap with server-side apply: unlike
# the hashicorp kubernetes provider's kubernetes_manifest, kubectl_manifest
# needs no definition at plan time, and the Stigmer's definition arrives only
# inside the chart's release. kubernetes writes the Secrets and the operator's
# namespace, and reads, at plan time, what the preflight checks.
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
  }
}

provider "kubernetes" {
}

provider "helm" {
}

provider "kubectl" {
}
