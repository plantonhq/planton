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

provider "google" {
  # The Identity Toolkit API requires a quota project on user-credential
  # calls: without this override, a deploy under plain ADC
  # (`gcloud auth application-default login`) fails at create with 403
  # "requires a quota project" (live-verified on the config kind — same
  # API serves tenants). The override attributes quota to the resource's
  # own project under every credential mode.
  user_project_override = true
}
