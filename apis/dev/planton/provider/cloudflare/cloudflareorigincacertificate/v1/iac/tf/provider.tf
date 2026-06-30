terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }
}

provider "cloudflare" {
  # Cloudflare provider automatically uses CLOUDFLARE_API_TOKEN environment variable
}
