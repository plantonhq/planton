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
      # Feature floor (from the provider changelog): the classic surface
      # (transition_default_minimum_object_size, partitioned log prefixes,
      # DSSE-KMS) predates 6.0, but this module also renders
      # aws_s3_bucket_metadata_configuration (6.5.0), the SSE rule's
      # blocked_encryption_types (6.22.0), aws_s3_bucket_abac (6.23.0), and
      # the bucket's bucket_namespace (6.37.0) — all comfortably under the
      # ~> 6.58 pin.
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
