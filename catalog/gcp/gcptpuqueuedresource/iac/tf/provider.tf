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
    # Google publishes queued TPU requests (`google_tpu_v2_queued_resource`) ONLY in the beta provider;
    # the GA provider has no code for them at the pin. That one resource
    # attaches `provider = google-beta` in main.tf under a recorded admission
    # in pkg/providerparity/admissions/google-beta.yaml (the admission guard
    # holds every attachment to an entry and every entry to an attachment).
    # The channel rides the same constraint as google so both resolve one
    # release; API enablement stays on the GA provider.
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 8.3"
    }
  }
}

provider "google" {
}

provider "google-beta" {
}
