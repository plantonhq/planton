terraform {
  required_providers {
    civo = {
      source  = "civo/civo"
      version = "~> 1.0"
    }
  }
}

provider "civo" {
  # Civo provider automatically uses CIVO_TOKEN environment variable
}
