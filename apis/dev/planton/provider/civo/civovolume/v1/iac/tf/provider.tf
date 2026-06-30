terraform {
  required_providers {
    civo = {
      source  = "civo/civo"
      version = ">= 1.0"
    }
  }
}

# Civo provider configuration
# Authentication is handled via CIVO_TOKEN environment variable
provider "civo" {
  region = var.spec.region
}

