terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region     = var.spec.region
  access_key = var.provider_config.access_key_id
  secret_key = var.provider_config.secret_access_key
  token      = var.provider_config.session_token != "" ? var.provider_config.session_token : null
}
