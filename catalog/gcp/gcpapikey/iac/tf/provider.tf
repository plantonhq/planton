terraform {
  required_providers {
    # Pessimistic minor float on the current major: every GCP module tracks the
    # same 8.x line so behavior is uniform across the catalog and upgrades to a
    # future major happen provider-wide in one decision, never per-kind. The
    # floor is the release the catalog is audited against, so the declared
    # constraint can never resolve below what any module's arguments need.
    # Every field this module uses is GA on the released 8.x line, so no
    # google-beta dependency exists to drift.
    google = {
      source  = "hashicorp/google"
      version = "~> 8.3"
    }
  }
}

provider "google" {
  # The API Keys API attributes quota to the caller's project on
  # user-credential calls: without this override, a deploy under plain ADC
  # (`gcloud auth application-default login`) fails at create with 403
  # "requires a quota project" -- the same behavior the Identity Toolkit API
  # shows, and the same fix. The override attributes quota to the key's own
  # project under every credential mode.
  user_project_override = true
}
