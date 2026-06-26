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
  # Cloudflare provider configuration.
  # API token is provided via the CLOUDFLARE_API_TOKEN environment variable.
}

# AWS provider aimed at the R2 S3-compatible endpoint, used only to fetch a
# pre-built worker bundle when spec.r2_bundle is set.
provider "aws" {
  alias                       = "r2"
  region                      = "auto"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_region_validation      = true
  skip_requesting_account_id  = true

  endpoints {
    s3 = "https://${var.spec.account_id}.r2.cloudflarestorage.com"
  }

  # R2 credentials are provided via AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY.
}
