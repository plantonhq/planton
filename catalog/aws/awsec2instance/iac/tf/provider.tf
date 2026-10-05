terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      # One pessimistic pin, catalog-wide: every AWS module tracks the same
      # provider line, floored at the newest minor already released when the
      # monthly pin sweep last advanced it. The `~>` cap makes the next major
      # a deliberate catalog-wide decision, and floor-at-latest-released-minor
      # means the constraint never understates what any module's newest
      # argument needs. Only the sweep moves this line — never a single kind.
      #
      # Feature floor: instance_market_options.market_type
      # "interruptible-capacity-reservation" needs the 6.58-era SDK enum
      # (the value rides the vendored AWS SDK, no dated changelog entry);
      # the capacity-block market's perpetual-diff fix landed in 6.53,
      # cpu_options.nested_virtualization in 6.33,
      # secondary_network_interface in 6.32, primary_network_interface in
      # 6.10, and force_destroy in 6.8.
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
