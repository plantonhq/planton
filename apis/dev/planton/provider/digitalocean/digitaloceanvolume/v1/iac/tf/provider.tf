terraform {
  required_version = ">= 1.5"

  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

# DigitalOcean provider configuration
# The token should be provided via environment variable DIGITALOCEAN_TOKEN
# or through Terraform variables/backend configuration
provider "digitalocean" {
  # Token is typically set via DIGITALOCEAN_TOKEN environment variable
  # or can be provided explicitly via a variable
}

