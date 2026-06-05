locals {
  cert_manager_namespace = var.spec.cert_manager_namespace
  dns_domain             = var.spec.dns_domain

  cloudflare_secret_name       = "${var.metadata.name}-cloudflare-credentials"
  acme_account_key_secret_name = "letsencrypt-${local.dns_domain}-account-key"

  is_cloudflare   = var.spec.cloudflare != null
  is_gcp          = var.spec.gcp_cloud_dns != null
  is_aws          = var.spec.aws_route53 != null
  is_azure        = var.spec.azure_dns != null

  # Build the DNS-01 solver for the active provider. Exactly one provider oneof
  # is set, so each branch contributes its single key (or nothing) and merge()
  # yields a properly typed object carrying only the active solver. This mirrors
  # kubernetescertificate's conditional-manifest-key pattern (iac/tf/main.tf).
  #
  # Do NOT collapse this back into a nested ternary returning differently-shaped
  # {dns01={...}} objects: Terraform unifies the disjoint branch types into a map,
  # which kubernetes_manifest then rejects on its strict map->object morph path
  # ("Failed to transform Map element into Object element type ... required
  # attribute apiKeySecretRef not set").
  solver = {
    dns01 = merge(
      local.is_cloudflare ? {
        cloudflare = {
          apiTokenSecretRef = {
            name = local.cloudflare_secret_name
            key  = "api-token"
          }
        }
      } : {},
      local.is_gcp ? {
        cloudDNS = {
          project = var.spec.gcp_cloud_dns.project_id
        }
      } : {},
      local.is_aws ? {
        route53 = {
          region = var.spec.aws_route53.region
        }
      } : {},
      local.is_azure ? {
        azureDNS = {
          subscriptionID    = var.spec.azure_dns.subscription_id
          resourceGroupName = var.spec.azure_dns.resource_group
        }
      } : {},
    )
  }
}
