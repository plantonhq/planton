# The hashicorp kubernetes provider renders the module-owned objects
# (namespace, env Secret, Roles, RoleBindings); helm installs the relay
# chart; kubectl (alekc/kubectl) applies the optional ServiceMonitor CR -
# unlike the hashicorp provider's kubernetes_manifest, kubectl_manifest
# needs no cluster connection at plan time, so the monitor plans before the
# Prometheus Operator's CRDs exist.
terraform {
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
