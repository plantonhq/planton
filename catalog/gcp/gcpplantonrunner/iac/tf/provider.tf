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
  # Project and credentials are injected by the runtime as environment
  # variables (GOOGLE_PROJECT + GOOGLE_CREDENTIALS or the ambient chain),
  # resolved from the stack input's provider_config. For keyless (oidc)
  # connections the runtime performs the web-identity exchange and injects
  # the resulting short-lived credentials. Keep this block empty -- do not
  # wire project or static keys here.
}
