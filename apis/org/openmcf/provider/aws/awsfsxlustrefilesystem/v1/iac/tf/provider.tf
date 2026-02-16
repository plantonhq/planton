# ---------------------------------------------------------------------------
# AWS Provider Configuration
# ---------------------------------------------------------------------------
# Provider credentials and region are supplied via variables to support
# dynamic credential injection (e.g., from stack inputs or CI/CD).
# ---------------------------------------------------------------------------

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "= 5.82.0"
    }
  }
}

provider "aws" {
  access_key = var.access_key
  secret_key = var.secret_key
  region     = var.region
  token      = var.session_token
}
