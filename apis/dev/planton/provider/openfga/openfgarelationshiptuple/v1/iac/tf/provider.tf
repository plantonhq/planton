# OpenFGA Provider Configuration
#
# The OpenFGA provider credentials are passed via environment variables:
# - FGA_API_URL (required): The OpenFGA server URL
# - FGA_API_TOKEN: For token-based authentication
# - FGA_CLIENT_ID, FGA_CLIENT_SECRET, FGA_API_TOKEN_ISSUER: For client credentials auth
#
# These environment variables are automatically configured by Planton
# from the OpenFgaProviderConfig in the stack input.
#
# Reference: https://registry.terraform.io/providers/openfga/openfga/latest/docs

terraform {
  required_version = ">= 1.0"

  required_providers {
    openfga = {
      source  = "openfga/openfga"
      version = ">= 0.5.0"
    }
  }
}

# Provider is configured via environment variables
provider "openfga" {}
