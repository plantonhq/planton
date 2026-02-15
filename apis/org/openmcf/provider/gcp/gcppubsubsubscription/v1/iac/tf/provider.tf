terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
    }
  }
}

provider "google" {
  credentials = var.provider_config.service_account_key_base64 != "" ? base64decode(var.provider_config.service_account_key_base64) : null
}
