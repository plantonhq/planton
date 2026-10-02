# Enable the Compute Engine API — the control plane that owns disks.
# disable_on_destroy is false: tearing down one disk must never disable
# the API for everything else in the project.
resource "google_project_service" "compute_api" {
  project = local.project_id
  service = "compute.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The persistent disk. Sharp edges, all taught by the API rather than
# invented here:
#
#   - name, zone, type, sources, encryption, and architecture are
#     ForceNew — changing them replaces the disk and its data. size grows
#     in place but never shrinks.
#
#   - At most one source (image / snapshot / instant snapshot / storage
#     object / source_disk); none creates an empty disk — the common case
#     for data volumes.
#
#   - Deleting a disk still attached to a running instance fails; detach
#     first (or delete the instance).
#
#   - create_snapshot_before_destroy takes a final snapshot during
#     destroy — a last-resort recovery net for precious volumes (CMEK
#     disks reuse their key for the snapshot).
resource "google_compute_disk" "this" {
  name    = local.disk_name
  project = local.project_id
  zone    = var.spec.zone
  labels  = local.final_labels

  description = var.spec.description != "" ? var.spec.description : null
  type        = local.type
  # Zero means unset: the disk then takes the image or snapshot size, or
  # the type's default, which the API computes.
  size = var.spec.size_gb != 0 ? var.spec.size_gb : null

  image                   = local.image
  snapshot                = local.snapshot
  source_instant_snapshot = local.source_instant_snapshot
  source_storage_object   = local.source_storage_object
  source_disk             = local.source_disk
  access_mode             = local.access_mode
  architecture            = local.architecture
  storage_pool            = local.storage_pool

  provisioned_iops       = var.spec.provisioned_iops
  provisioned_throughput = var.spec.provisioned_throughput

  enable_confidential_compute = var.spec.enable_confidential_compute ? true : null
  physical_block_size_bytes   = var.spec.physical_block_size_bytes

  create_snapshot_before_destroy        = var.spec.create_snapshot_before_destroy
  create_snapshot_before_destroy_prefix = var.spec.snapshot_before_destroy_prefix != "" ? var.spec.snapshot_before_destroy_prefix : null

  # Guest OS features for bootable disks; licenses for BYOL imports.
  # Both create-time only.
  dynamic "guest_os_features" {
    for_each = var.spec.guest_os_features
    content {
      type = guest_os_features.value
    }
  }
  licenses = length(var.spec.licenses) > 0 ? var.spec.licenses : null

  # Destroy-time guard: PREVENT fails the destroy; ABANDON unmanages the
  # disk without deleting it. Null falls back to the provider default
  # (DELETE).
  deletion_policy = local.deletion_policy

  # CMEK: the Compute Engine service agent must hold
  # roles/cloudkms.cryptoKeyEncrypterDecrypter on this key before create.
  # kms_key_service_account overrides which identity performs the
  # encryption request. Raw CSEK keys are deliberately not supported —
  # key material never flows through manifests or state.
  dynamic "disk_encryption_key" {
    for_each = var.spec.kms_key != "" ? [var.spec.kms_key] : []
    content {
      kms_key_self_link       = disk_encryption_key.value
      kms_key_service_account = var.spec.kms_key_service_account != "" ? var.spec.kms_key_service_account : null
    }
  }

  # CMEK decryption of an encrypted source image / snapshot.
  dynamic "source_image_encryption_key" {
    for_each = var.spec.source_image_encryption != null ? [var.spec.source_image_encryption] : []
    content {
      kms_key_self_link       = source_image_encryption_key.value.kms_key
      kms_key_service_account = source_image_encryption_key.value.kms_key_service_account != "" ? source_image_encryption_key.value.kms_key_service_account : null
    }
  }
  dynamic "source_snapshot_encryption_key" {
    for_each = var.spec.source_snapshot_encryption != null ? [var.spec.source_snapshot_encryption] : []
    content {
      kms_key_self_link       = source_snapshot_encryption_key.value.kms_key
      kms_key_service_account = source_snapshot_encryption_key.value.kms_key_service_account != "" ? source_snapshot_encryption_key.value.kms_key_service_account : null
    }
  }

  # Async replication: pairing this disk (the SECONDARY) to its primary.
  # Replication starts when the pair is activated on the primary.
  dynamic "async_primary_disk" {
    for_each = local.async_primary_disk != null ? [local.async_primary_disk] : []
    content {
      disk = async_primary_disk.value
    }
  }

  # Resource Manager tags bind at create time only.
  dynamic "params" {
    for_each = length(var.spec.resource_manager_tags) > 0 ? [var.spec.resource_manager_tags] : []
    content {
      resource_manager_tags = params.value
    }
  }

  depends_on = [
    google_project_service.compute_api,
  ]
}
