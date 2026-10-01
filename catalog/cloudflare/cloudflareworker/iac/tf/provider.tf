terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.23"
    }
    # Reads spec.r2_bundle through R2's S3-compatible API. 6.29 is the floor for
    # aws_s3_object's download_body, which returns the object's bytes whatever its
    # declared Content-Type (see main.tf).
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.58"
    }
  }
}

provider "cloudflare" {
  # Cloudflare provider configuration.
  # API token is provided via the CLOUDFLARE_API_TOKEN environment variable.
}

# An S3-compatible provider aimed at the account's R2 endpoint. Only the bundle read
# (data.aws_s3_object.bundle) uses it, and only when spec.r2_bundle is set.
#
# OpenTofu and Terraform configure every provider a module references, even when the
# only block that uses it has count = 0, and this provider refuses to configure without
# SOME credential source even with every skip flag set. A Worker with inline content
# must deploy with nothing but its Cloudflare credential, so without a bundle the
# provider gets a fixed placeholder key pair: with the skip flags below it makes no
# network call while configuring, and with no block reading through it the placeholder
# never signs a request. With a bundle it gets the connection's R2 pair (credentials.tf);
# a null pair there defers to the provider's own credential chain.
#
# A provider-level for_each would avoid the placeholder, but Terraform cannot load a
# module that uses one, and this module serves both engines.
provider "aws" {
  alias      = "r2"
  region     = "auto"
  access_key = local.use_bundle ? var.r2_access_key_id : "unused-no-r2-bundle"
  secret_key = local.use_bundle ? var.r2_secret_access_key : "unused-no-r2-bundle"

  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_region_validation      = true
  skip_requesting_account_id  = true
  s3_use_path_style           = true

  endpoints {
    s3 = local.r2_endpoint
  }
}
