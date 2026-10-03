terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      # One pessimistic pin, catalog-wide: every AWS module tracks the same
      # provider line, floored at the newest minor already released when the
      # monthly pin sweep last advanced it. The `~>` cap makes the next major
      # a deliberate catalog-wide decision, and floor-at-latest-released-minor
      # means the constraint never understates what any module's newest
      # argument needs. Only the sweep moves this line — never a single kind.
      #
      # Feature floor: the scraper's VPC source arm, role_configuration,
      # and aws_prometheus_scraper_logging_configuration are 6.5x-era
      # surface - the catalog pin is exactly their floor generation.
      source  = "hashicorp/aws"
      version = "~> 6.58"
    }
  }
}

provider "aws" {
  # Region and credentials are injected by the runtime as environment variables
  # (AWS_REGION + AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY / AWS_SESSION_TOKEN), resolved
  # from the IaC input's provider_config. For keyless (oidc)
  # connections the runtime performs the STS web-identity exchange and injects the resulting
  # short-lived credentials. Keep this block empty -- do not wire region or static keys here.
}

# AWS publishes the default scrape configuration for EKS-sourced
# scrapers; the module resolves it at deploy when the spec leaves
# scrape_configuration unset (spec-guaranteed to happen only on the EKS
# arm).
data "aws_prometheus_default_scraper_configuration" "this" {
  count = var.spec.scrape_configuration == "" ? 1 : 0
}
