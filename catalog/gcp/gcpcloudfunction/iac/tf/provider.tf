terraform {
  required_providers {
    # Float on the 8.x line: patch/minor provider fixes arrive without a
    # per-kind pin edit, and moves to a future major happen provider-wide in
    # one decision, never per-kind. Every field this module uses is GA on
    # the released 8.x line, so no google-beta dependency exists to drift.
    google = {
      source  = "hashicorp/google"
      version = "~> 8.3"
    }
  }
}

provider "google" {
}
