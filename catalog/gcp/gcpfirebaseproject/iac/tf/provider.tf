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
    # Google publishes Firebase's core resources (the project enablement, the
    # default storage bucket) ONLY in the beta provider; the GA provider has
    # no code for them at the pin. Those two resources -- and only those --
    # attach `provider = google-beta` below, each under a recorded admission
    # in pkg/providerparity/admissions/google-beta.yaml (the admission guard
    # holds every attachment to an entry and every entry to an attachment).
    # The channel rides the same constraint as google so both resolve one
    # release; App Check and API enablement stay on the GA provider.
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 8.3"
    }
  }
}

# The Firebase Management API attributes quota to the caller's project on
# user-credential calls -- Google's own Firebase provider docs say to set
# user_project_override, and under plain ADC (`gcloud auth
# application-default login`) a create without it fails with 403 "requires
# a quota project". The override attributes quota to the project being
# enabled under every credential mode. Both provider blocks carry it so
# behavior never depends on which channel a resource rides.
# billing_project NAMES the quota project beside the override: the override attributes a
# resource call to the resource's own project, but a data-source read carries no project the
# header can borrow, so under a user credential Google attributes it to its shared ADC project
# (API disabled there) and the read fails with 403 "requires a quota project" after the create
# succeeded (live-verified 2026-09-18). Service-account and keyless credentials are unaffected.
provider "google" {
  user_project_override = true
  billing_project       = local.project_id
}

provider "google-beta" {
  user_project_override = true
  billing_project       = local.project_id
}
