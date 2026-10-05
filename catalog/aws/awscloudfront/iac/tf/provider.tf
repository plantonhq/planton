terraform {
  required_providers {
    aws = {
      # One pessimistic pin, catalog-wide: every AWS module tracks the same
      # provider line, floored at the newest minor already released when the
      # monthly pin sweep last advanced it. The `~>` cap makes the next major
      # a deliberate catalog-wide decision, and floor-at-latest-released-minor
      # means the constraint never understates what any module's newest
      # argument needs. Only the sweep moves this line — never a single kind.
      #
      # Feature floor 6.51.0: origin_mtls_config (origin-side mTLS) landed there.
      # Earlier floors this module renders: cache_tag_config 6.46.0;
      # connection_function_association + viewer_mtls_config +
      # vpc_origin_config.owner_account_id 6.28.0; custom_origin_config.ip_address_type
      # 6.15.0; origin.response_completion_timeout 6.13.0; anycast_ip_list_id 6.3.0.
      # Everything else (incl. continuous-deployment policies, staging, vpc_origin_config,
      # and per-behavior grpc_config) predates 6.0.
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
