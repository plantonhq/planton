terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0, < 6.0"
    }
  }
  
  required_version = ">= 1.0"
}

# AWS provider configuration
# Region and other settings should be configured by the caller
provider "aws" {
  region = var.spec.region
}


