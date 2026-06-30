# Civo Provider Configuration
#
# The Civo provider requires an API token for authentication.
# This should be provided via the CIVO_TOKEN environment variable.
#
# Example:
#   export CIVO_TOKEN="your-civo-api-token"
#   terraform init
#   terraform apply

terraform {
  required_providers {
    civo = {
      source  = "civo/civo"
      version = "~> 1.1"
    }
  }
}

provider "civo" {
  # The provider will automatically use the CIVO_TOKEN environment variable
  # for authentication. No explicit configuration needed here.
  
  # Optional: You can specify the region here, but it's better to specify
  # it per-resource for flexibility
  # region = var.spec.region
}

