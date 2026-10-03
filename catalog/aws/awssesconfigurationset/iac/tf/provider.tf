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
      # Feature floor (v6 family baseline): the full SESv2 configuration-set surface modeled
      # here (including tracking https_policy and delivery
      # max_delivery_seconds, both provider v5.81.0+) is stable before the
      # v6 line; the floor keeps the SES family on one provider generation.
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
