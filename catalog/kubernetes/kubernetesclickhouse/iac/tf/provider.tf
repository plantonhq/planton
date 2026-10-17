# Provider requirements for the KubernetesClickHouse module.
#
# kubectl (alekc/kubectl) applies the ClickHouseInstallation and
# ClickHouseKeeperInstallation custom resources: unlike the hashicorp
# kubernetes provider's kubernetes_manifest resource, kubectl_manifest
# needs no cluster connection at plan time, so the cluster can be planned
# before the Altinity operator's CRDs exist (single-run infra charts,
# offline plan proofs). The hashicorp kubernetes provider materializes the
# optional namespace and the auth Secret; the random provider generates the
# password of every user declared without one.
#
# Both providers are configured by the calling workspace/environment (the
# same kubeconfig environment contract).

terraform {
  required_version = ">= 1.0"

  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.20"
    }
    kubectl = {
      source  = "alekc/kubectl"
      version = ">= 2.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }
}

provider "kubernetes" {
}

provider "kubectl" {
}
