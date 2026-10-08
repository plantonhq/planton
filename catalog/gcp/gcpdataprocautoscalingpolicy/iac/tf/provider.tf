terraform {
  required_providers {
    # Pessimistic minor float on the current major: every GCP module tracks the
    # same 8.x line so behavior is uniform across the catalog and upgrades to a
    # future major happen provider-wide in one decision, never per-kind. The
    # floor is the release the catalog is audited against, so the declared
    # constraint can never resolve below what any module's arguments need. Every
    # field this module sends is GA on the released 8.x line, so no google-beta
    # dependency exists to drift.
    google = {
      source  = "hashicorp/google"
      version = "~> 8.3"
    }
  }
}

provider "google" {
  # The project this component's spec names (null when it names none). An
  # import records it, so a taken-over resource is never planned for
  # replacement. Credentials are injected by the runtime; the connection
  # never names a project.
  project = local.project_id
}
