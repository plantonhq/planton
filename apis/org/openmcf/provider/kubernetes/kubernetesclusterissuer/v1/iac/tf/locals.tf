locals {
  cert_manager_namespace = var.spec.cert_manager_namespace
  dns_domain             = var.spec.dns_domain

  cloudflare_secret_name       = "${var.metadata.name}-cloudflare-credentials"
  acme_account_key_secret_name = "letsencrypt-${local.dns_domain}-account-key"

  is_cloudflare   = var.spec.cloudflare != null
  is_gcp          = var.spec.gcp_cloud_dns != null
  is_aws          = var.spec.aws_route53 != null
  is_azure        = var.spec.azure_dns != null

  solver = (
    local.is_gcp ? {
      dns01 = {
        cloudDNS = {
          project = var.spec.gcp_cloud_dns.project_id
        }
      }
    } :
    local.is_aws ? {
      dns01 = {
        route53 = {
          region = var.spec.aws_route53.region
        }
      }
    } :
    local.is_azure ? {
      dns01 = {
        azureDNS = {
          subscriptionID    = var.spec.azure_dns.subscription_id
          resourceGroupName = var.spec.azure_dns.resource_group
        }
      }
    } :
    local.is_cloudflare ? {
      dns01 = {
        cloudflare = {
          apiTokenSecretRef = {
            name = local.cloudflare_secret_name
            key  = "api-token"
          }
        }
      }
    } : null
  )
}
