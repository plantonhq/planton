# The Cloudflare Pages project: build config, optional git source, and
# per-environment deployment configuration.
resource "cloudflare_pages_project" "main" {
  account_id        = var.spec.account_id
  name              = var.spec.name
  production_branch  = var.spec.production_branch

  build_config       = local.build_config
  source             = local.source
  deployment_configs = local.deployment_configs
}

# Custom domains attached to the project (each a hostname in a zone on this account).
resource "cloudflare_pages_domain" "main" {
  for_each = local.domains_map

  account_id   = var.spec.account_id
  project_name = cloudflare_pages_project.main.name
  name         = each.value
}
