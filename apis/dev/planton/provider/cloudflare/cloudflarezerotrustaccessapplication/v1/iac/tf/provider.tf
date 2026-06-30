# provider.tf

terraform {
  required_version = ">= 1.0"

  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.0"
    }
  }
}

# Cloudflare provider configuration
# API token should be set via environment variable: CLOUDFLARE_API_TOKEN
provider "cloudflare" {
  # Configuration from environment variables or Terraform Cloud
}

