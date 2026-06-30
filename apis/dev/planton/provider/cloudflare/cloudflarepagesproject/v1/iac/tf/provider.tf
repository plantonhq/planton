terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.0"
    }
  }
}

provider "cloudflare" {
  # Cloudflare provider configuration.
  # API token is provided via the CLOUDFLARE_API_TOKEN environment variable.
}
