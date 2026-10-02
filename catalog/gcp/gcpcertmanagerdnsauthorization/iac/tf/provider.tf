terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 8.3"
    }
  }
}

provider "google" {
}
