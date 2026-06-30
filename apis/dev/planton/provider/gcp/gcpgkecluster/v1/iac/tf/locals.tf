locals {
  # Create GCP labels from metadata and spec
  gcp_labels = {
    resource     = var.spec.cluster_name
    env          = var.metadata.env.id
    resource-id  = var.metadata.id
    resource-org = var.metadata.org
  }

  # Map proto enum to GKE release channel string
  # 0=unspecified, 1=RAPID, 2=REGULAR, 3=STABLE, 4=NONE
  release_channel_map = {
    0 = "REGULAR"      # Default to REGULAR if unspecified
    1 = "RAPID"
    2 = "REGULAR"
    3 = "STABLE"
    4 = "UNSPECIFIED"  # NONE in proto means manual upgrades (UNSPECIFIED in GKE API)
  }
  
  release_channel = lookup(local.release_channel_map, var.spec.release_channel, "REGULAR")
}

