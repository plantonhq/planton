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
      # Feature floor: the newest surface this module renders is the
      # aws_lambda_function_scaling_config resource (6.54.0). The rest of
      # the modern surface arrived across the 6.2x line: tenancy_config
      # 6.23.0, capacity_provider_config + publish_to 6.24.0,
      # durable_config 6.25.0. The classic surface (logging_config,
      # snap_start, ipv6_allowed_for_dual_stack, source_kms_key_arn, the
      # recursion and runtime-management resources) predates v6.
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
