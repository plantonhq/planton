terraform {
  required_version = ">= 1.0"

  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

# DigitalOcean provider configuration is expected to be provided by the caller
# through environment variables (DIGITALOCEAN_TOKEN) or other Terraform configuration methods.
# The token can also be set via:
#   - Environment variable: export DIGITALOCEAN_TOKEN="your-token-here"
#   - Terraform CLI: terraform apply -var="do_token=your-token-here"
#   - terraform.tfvars file

