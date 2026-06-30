locals {
  # Create GCP labels from metadata
  gcp_labels = {
    resource     = var.metadata.name
    env          = var.metadata.env.id
    resource-id  = var.metadata.id
    resource-org = var.metadata.org
  }
}

