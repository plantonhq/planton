# Auth0 Provider Configuration
# Documentation: https://registry.terraform.io/providers/auth0/auth0/latest/docs

variable "auth0_credential" {
  description = "Auth0 API authentication credentials"
  type = object({
    domain        = string
    client_id     = string
    client_secret = string
  })
  sensitive = true
}

terraform {
  required_version = ">= 1.0"

  required_providers {
    auth0 = {
      source  = "auth0/auth0"
      version = "~> 1.0"
    }
  }
}

provider "auth0" {
  domain        = var.auth0_credential.domain
  client_id     = var.auth0_credential.client_id
  client_secret = var.auth0_credential.client_secret
}
