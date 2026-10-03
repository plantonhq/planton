terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 5.0"
    }
  }

  required_version = ">= 1.0"
}

provider "azurerm" {
  features {}
  # Subscription and credentials are injected by the runtime as environment
  # variables (ARM_SUBSCRIPTION_ID + ARM_CLIENT_ID / ARM_CLIENT_SECRET or
  # the keyless web-identity exchange), resolved from the IaC input's
  # provider_config. Keep this block bare -- do not wire credentials here.
}
