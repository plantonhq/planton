terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
    }
  }
}

# Provider configuration is expected to be passed from the root module.
provider "aws" {
  region = var.spec.region
}


