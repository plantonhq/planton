terraform {
  required_providers {
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
