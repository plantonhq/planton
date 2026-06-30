terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.35"
    }
    # If you need helm or other providers, declare them here as well.
  }
}

provider "kubernetes" {
}
