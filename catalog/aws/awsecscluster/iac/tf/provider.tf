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
      # Feature floor: the cluster's managed_storage_configuration and the
      # capacity provider's managed_draining land on the v6 line; the ECS
      # Managed Instances family (managed_instances_provider + the cluster
      # binding) lands in 6.15.0 and is completed by 6.58.0
      # (capacity_reservations, the RESERVED capacity_option_type, and
      # local_storage_configuration) -- this pin is the first release
      # carrying the full managed-instances surface this module drives.
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
