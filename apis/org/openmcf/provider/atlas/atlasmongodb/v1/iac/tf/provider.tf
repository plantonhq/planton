# Atlas MongoDB Provider Configuration
# This file configures the Atlas MongoDB Terraform provider
# Documentation: https://registry.terraform.io/providers/mongodb/atlasmongodb/latest/docs

# Variables for Atlas MongoDB authentication
variable "atlasmongodb_credential" {
  description = "Atlas MongoDB authentication credentials"
  type = object({
    # Atlas MongoDB Public API Key
    # Create API keys in Atlas UI: Project Settings -> Access Manager -> API Keys
    public_key = string

    # Atlas MongoDB Private API Key
    # This key is shown only once when created and should be stored securely
    private_key = string
  })
  sensitive = true
}

# Configure the Atlas MongoDB Provider
terraform {
  required_providers {
    atlasmongodb = {
      source  = "mongodb/atlasmongodb"
      version = "~> 1.14"
    }
  }
}

# Provider configuration
# The provider uses the public and private keys for API authentication
provider "atlasmongodb" {
  public_key  = var.atlasmongodb_credential.public_key
  private_key = var.atlasmongodb_credential.private_key
}

