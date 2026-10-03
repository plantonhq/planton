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
      # Feature floor: the v6 line carries the full
      # aws_redshiftserverless_workgroup surface this module uses,
      # including price_performance_target and track_name, plus the
      # satellites this module also renders
      # (aws_redshiftserverless_endpoint_access, _usage_limit,
      # _custom_domain_association) -- all predate the v6 line.
      source  = "hashicorp/aws"
      version = "~> 6.58"
    }
    time = {
      # Utility provider for the destroy-side settle between the
      # usage-limit deletes and the endpoint-access delete (see
      # satellite_settle.tf): Redshift Serverless holds a per-workgroup
      # operation lock for ~15-30s AFTER a usage-limit call returns, and
      # DeleteEndpointAccess conflicts against it with no provider-side
      # retry (the aws provider retries that ConflictException only on
      # the workgroup's own delete/update).
      source  = "hashicorp/time"
      version = "~> 0.13"
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
