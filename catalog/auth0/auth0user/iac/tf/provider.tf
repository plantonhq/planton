terraform {
  required_version = ">= 1.0"

  required_providers {
    auth0 = {
      source  = "auth0/auth0"
      version = "~> 1.0"
    }
    # random mints the initial password when the spec declares none on a
    # database connection (the credential a module can mint is never asked
    # of the person). Twin: the Pulumi module's pulumi-random dependency.
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }
}
