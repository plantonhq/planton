terraform {
  required_version = ">= 1.0"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
}

provider "azurerm" {
  features {}

  subscription_id = var.provider_config.subscription_id
  tenant_id       = var.provider_config.tenant_id
  client_id       = var.provider_config.client_id
  client_secret   = var.provider_config.client_secret
}
