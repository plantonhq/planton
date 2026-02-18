terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
    }
  }
}

provider "aws" {
  region     = var.spec.region
  access_key = var.provider_config.access_key_id != "" ? var.provider_config.access_key_id : null
  secret_key = var.provider_config.secret_access_key != "" ? var.provider_config.secret_access_key : null
  token      = var.provider_config.session_token != "" ? var.provider_config.session_token : null
}
