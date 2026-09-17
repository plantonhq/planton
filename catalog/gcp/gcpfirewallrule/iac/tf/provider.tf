terraform {
  required_providers {
    # Pessimistic minor float on the current major: every GCP module tracks the
    # same 8.x line so behavior is uniform across the catalog and upgrades to a
    # future major happen provider-wide in one decision, never per-kind. The
    # floor is the release the catalog is audited against, so the declared
    # constraint can never resolve below what any module's arguments need.
    google = {
      source  = "hashicorp/google"
      version = "~> 8.3"
    }
  }
}

# Empty ambient block: the runner injects credentials through the standard
# GOOGLE_* environment chain, never through module variables.
provider "google" {
}
