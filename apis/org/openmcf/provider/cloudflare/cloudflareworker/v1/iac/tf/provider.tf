terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.0"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "cloudflare" {
  # Cloudflare provider configuration
  # API token should be provided via CLOUDFLARE_API_TOKEN environment variable
}

# AWS provider configured for R2 access
# Required to fetch worker bundle from R2 bucket
provider "aws" {
  alias                       = "r2"
  region                      = "auto"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_region_validation      = true
  skip_requesting_account_id  = true

  # R2 endpoint
  endpoints {
    s3 = "https://${var.spec.account_id}.r2.cloudflarestorage.com"
  }

  # R2 credentials should be provided via:
  # AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables
}

