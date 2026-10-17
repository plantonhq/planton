# The hashicorp kubernetes provider renders the module-owned objects (the
# namespace and the instance Secret); helm installs the Langfuse chart; the
# random provider generates every credential the spec leaves empty.
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
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }
}

provider "kubernetes" {
}

provider "helm" {
}
