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
    # Google publishes the service-agent create (`google_project_service_identity`)
    # ONLY in the beta provider; the GA provider has no such resource at the
    # pin. That one resource attaches `provider = google-beta` in main.tf under
    # a recorded admission in pkg/providerparity/admissions/google-beta.yaml.
    # The channel rides the same constraint as google so both resolve one
    # release; everything else stays on the GA provider.
    google-beta = {
      source  = "hashicorp/google-beta"
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

provider "google-beta" {
  # The project this component's spec names (null when it names none), so the
  # service agent is created in the same project as the issuance config.
  # Credentials are injected by the runtime.
  project = local.project_id
}
