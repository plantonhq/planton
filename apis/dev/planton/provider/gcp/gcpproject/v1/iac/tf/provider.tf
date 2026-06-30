terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "6.19.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }
}

provider "google" {
}

provider "random" {
}
