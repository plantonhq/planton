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
      # Feature floor: aws_bedrock_guardrail's cross_region_config +
      # tier_config land in v6.6.0, the PII/regex action-enabled arms in
      # v6.8.0, the word-policy arms in v6.13.0, the content-filter
      # action/enabled/modalities arms in v6.22.0, the 1000-char topic
      # definitions in v6.42.0 -- and v6.54.0 fixes "inconsistent result"
      # errors when adding content/topic policy blocks, so an older floor
      # would ship known plan-time bugs. The guardrail_version resource
      # predates 6.x.
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
