terraform {
  required_providers {
    # Float on the 8.x line: patch/minor provider fixes arrive without a
    # per-kind pin edit, and moves to a future major happen provider-wide in
    # one decision, never per-kind. Every field this module uses is GA on
    # the released 8.x line (Cloud Run v2, Secret Manager, IAM), so no
    # google-beta dependency exists to drift.
    google = {
      source  = "hashicorp/google"
      version = "~> 8.3"
    }
  }

  required_version = ">= 1.0"
}

provider "google" {
  # The project this component's spec names (null when it names none). An
  # import records it, so a taken-over resource is never planned for
  # replacement. Credentials are injected by the runtime; the connection
  # never names a project.
  project = local.project_id
}
